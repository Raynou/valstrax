## Context

El bot vive hoy en un único `main.go`: abre una conexión WebSocket al gateway de Discord con `discordgo`, registra los comandos al arrancar y sirve seis slash commands contra un SQLite en `/tmp/peliculas.db`. El proceso tiene que estar encendido permanentemente y la ruta `/tmp` no garantiza durabilidad.

La observación que habilita este cambio: el bot **nunca** registra handlers de eventos del gateway. Su único `AddHandler` atiende `InteractionCreate`, es decir, exactamente lo que Discord también entrega por HTTP a un Interactions Endpoint. La conexión permanente no está comprando nada.

Restricciones de plataforma que acotan el diseño:
- Workers Free: 100k peticiones/día, 10 ms de CPU por petición, bundle de 3 MB comprimido, 50 subpeticiones por invocación.
- D1 Free: 5M filas leídas/día, 100k escritas/día, 50 consultas por invocación, 500 MB por base.
- Discord: la respuesta inicial a una interacción debe salir en menos de 3 segundos, y el endpoint debe superar una validación firmada antes de ser aceptado.

Ninguno de esos techos está cerca del volumen real (un servidor de amigos, decenas de interacciones al día).

## Goals / Non-Goals

**Goals:**
- Eliminar el proceso permanente y la máquina que lo aloja.
- Garantizar que el catálogo sobreviva reinicios y despliegues.
- Preservar exactamente la superficie de comandos actual y el aislamiento por `guild_id`.
- Quedarse dentro de la capa gratuita sin vigilancia activa.

**Non-Goals:**
- Añadir comandos o cambiar el comportamiento visible de los existentes.
- Paginar `/lista` (sigue con su `LIMIT 25` y el TODO heredado).
- Soportar eventos del gateway. Se renuncia a ellos de forma consciente y permanente mientras dure esta arquitectura.
- Mantener el código Go operativo en paralelo. No hay despliegue dual.

## Decisions

### D1: Interactions HTTP en lugar del gateway

Discord ofrece dos formas de recibir interacciones: gateway (WebSocket persistente) o Interactions Endpoint (HTTP). El gateway exige un proceso vivo; el HTTP no. Como el bot solo consume `InteractionCreate`, HTTP cubre el 100% del caso de uso.

**Alternativa descartada:** mantener el gateway en un VPS o en Oracle Always Free. Es más barato en esfuerzo (una línea: mover el fichero SQLite fuera de `/tmp`) pero conserva una máquina que parchear y un proceso que se cae. Se descartó por preferencia explícita de no mantener infraestructura.

**Coste aceptado:** cerrar la puerta a funciones basadas en eventos. Volver atrás significa rehacer el transporte.

### D2: TypeScript en lugar de Go sobre TinyGo

Workers ejecuta JS o WASM, no binarios nativos. Portar Go implicaría TinyGo con `syumai/workers`, y ahí se acumulan los problemas: un binario Go completo no cabe en los 3 MB comprimidos del plan gratuito (los propios ejemplos del proyecto usan TinyGo por eso), `discordgo` arrastra cliente HTTP, WebSocket y `encoding/json` intensivo en reflexión, y el conector D1 está en alpha.

El worker real son ~6 handlers, cada uno una consulta y un JSON de respuesta. En TypeScript eso es código directo sin sorpresas de toolchain.

**Alternativa descartada:** Cloud Run con el Go actual intacto. Conserva `discordgo` y escala a cero, pero reintroduce contenedor y registry. Se descartó porque el escalado a cero no resuelve ningún problema que este bot tenga.

### D3: Verificación Ed25519 con WebCrypto nativo

Workers soporta Ed25519 en WebCrypto, así que la verificación no necesita dependencias. Se prefiere sobre el paquete `discord-interactions` para mantener el bundle mínimo y el árbol de dependencias en cero.

**Detalle crítico de implementación:** la firma se calcula sobre el cuerpo **crudo** concatenado al timestamp. Hay que leer `await request.text()` y verificar *antes* de parsear JSON; parsear y re-serializar rompe la verificación de forma intermitente y difícil de diagnosticar.

### D4: Respuesta síncrona, sin deferred

Discord permite responder `DEFERRED_CHANNEL_MESSAGE_WITH_SOURCE` y completar después vía webhook. No se usa: las consultas son triviales y una respuesta directa cabe holgadamente en los 3 segundos. Diferir añadiría una segunda petición saliente y estado intermedio sin ganar nada.

### D5: D1 con el esquema actual sin rediseñar

D1 es SQLite, así que la tabla `peliculas` se traslada tal cual — mismas columnas, mismo índice `idx_peliculas_guild`, mismos `COLLATE NOCASE`. Las consultas se copian casi literalmente; solo cambia la API que las ejecuta (`db.prepare(...).bind(...)`).

**Consecuencia:** la capability `guild-scoped-movies` no cambia de requisitos. Su spec queda intacta.

### D6: Registro de comandos como script one-off

Sin proceso que arranque, el `ApplicationCommandBulkOverwrite` no tiene dónde vivir. Pasa a un script ejecutable a mano (`npm run register`) que llama a la REST API de Discord con el token. Se ejecuta solo cuando cambia la definición de comandos, no en cada despliegue.

### D7: Secretos como Workers secrets

`DISCORD_TOKEN`, `DISCORD_PUBLIC_KEY` y `DISCORD_APPLICATION_ID` pasan de `.env` a `wrangler secret put`. El `.env` actual deja de usarse en producción.

## Risks / Trade-offs

- **La verificación de firma se implementa sobre el cuerpo ya parseado** → Falla de forma intermitente y confusa. Mitigación: leer el cuerpo crudo una sola vez con `request.text()`, verificar, y solo entonces `JSON.parse`. Cubierto por escenario explícito en la spec.

- **Discord rechaza el Interactions Endpoint al configurarlo** → El bot queda inutilizable hasta arreglarlo. Discord valida la URL enviando una petición `PING` firmada, y una firma mal verificada o un `PONG` mal formado hacen fallar el guardado. Mitigación: desplegar y verificar el handshake antes de tocar la configuración en el Developer Portal.

- **Rollback no es simétrico** → Volver a Go significa reimportar de D1 a SQLite. Mitigación: conservar el `.dump` original y el commit del `main.go` como punto de retorno durante las primeras semanas.

- **El límite de 10 ms de CPU** → Riesgo bajo: solo cuenta CPU real, no la espera de D1, y la verificación Ed25519 nativa es sub-milisegundo. Se anota por si en el futuro se añade lógica pesada.

## Migration Plan

1. Crear la base D1 y aplicar el esquema.
2. Importar el dump con `wrangler d1 execute`.
3. Desplegar el worker y verificar el handshake `PING`/`PONG` contra la URL desplegada.
4. Fijar el Interactions Endpoint URL en el Developer Portal y confirmar que Discord lo acepta.
5. Ejecutar el script de registro de comandos.
6. Probar los 6 comandos y los 2 autocompletados en un servidor real.
7. Retirar el código Go del repositorio.

**Rollback:** el binario en Go no esta siendo ejecutado actualmente, por lo que el plan para revertir esto es simplemente borrar todo lo creado en cloudfare y volver al ultimo commit donde estaba Go

## Open Questions

- ¿El código Go se borra del repositorio o se conserva en una rama/tag como referencia? Se puede conservar en una rama
legacy sin problema
- ¿Se mantiene la variable `GUILD_ID` para registrar comandos en un solo servidor durante pruebas, o se registran siempre globales?
