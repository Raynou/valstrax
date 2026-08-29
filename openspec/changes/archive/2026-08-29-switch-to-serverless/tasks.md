## 1. Scaffolding del proyecto Workers

- [x] 1.1 Inicializar el proyecto TypeScript de Workers (`package.json`, `tsconfig.json`, `wrangler.toml`) en la raíz del repo
- [x] 1.2 Añadir `wrangler` y `@cloudflare/workers-types` como dependencias de desarrollo
- [x] 1.3 Crear la base D1 (`wrangler d1 create valstrax-db`) y declarar su binding en `wrangler.toml`
- [x] 1.4 Registrar `DISCORD_PUBLIC_KEY` con `wrangler secret put` (único secreto que el worker
      lee en runtime) y replicarlo en `.dev.vars` para `wrangler dev`. `DISCORD_TOKEN` y
      `DISCORD_APPLICATION_ID` se quedan en `.env`: sólo los usa `scripts/register.ts`, que
      corre en local y nunca se empaqueta en el worker
- [x] 1.5 Verificar que `wrangler dev` arranca y responde a una petición vacía

## 2. Verificación de firma y enrutado

- [x] 2.1 Implementar `verifyRequest(request, publicKey)` con WebCrypto Ed25519, leyendo el cuerpo con `request.text()` una sola vez y verificando `timestamp + body` antes de parsear
- [x] 2.2 Devolver HTTP 401 cuando la firma sea inválida o falte alguna de las cabeceras `X-Signature-Ed25519` / `X-Signature-Timestamp`, sin tocar la base de datos
- [x] 2.3 Responder `{"type": 1}` a las interacciones de tipo `PING` (tipo 1)
- [x] 2.4 Implementar el router que despacha por tipo de interacción: comando (2) al handler por nombre, autocompletado (4) al proveedor de sugerencias
- [x] 2.5 Devolver un mensaje de error controlado ante un nombre de comando sin handler registrado
- [x] 2.6 Añadir la guardia de `guild_id` ausente (interacciones en DM) que responde el mensaje de error actual

## 3. Capa de datos sobre D1

- [x] 3.1 Escribir el fichero de esquema con `CREATE TABLE peliculas` (columnas `id`, `nombre`, `agregada_por`, `guild_id`, `creada_en`, `vista_en`) y `CREATE INDEX idx_peliculas_guild`
- [x] 3.2 Aplicar el esquema a la base D1 local y a la remota con `wrangler d1 execute`
- [x] 3.3 Crear el módulo de acceso a datos que envuelve `env.DB.prepare(...).bind(...)`, con una función por consulta portada de `main.go`
- [x] 3.4 Verificar que todas las consultas del módulo incluyen el filtro por `guild_id`

## 4. Handlers de comandos

- [x] 4.1 Portar `agregar`: INSERT con `guild_id` y respuesta con el id generado
- [x] 4.2 Portar `sugerir`: SELECT aleatorio de no vistas del guild, con el mensaje de catálogo agotado cuando no haya filas
- [x] 4.3 Portar `vista`: SELECT por nombre `COLLATE NOCASE` + UPDATE de `vista_en`, ambos filtrados por `guild_id`
- [x] 4.4 Portar `lista`: SELECT con `LIMIT 25`, marcadores ⬜/✅ y mensaje de catálogo vacío
- [x] 4.5 Portar `quitar`: DELETE por nombre o por id, exigiendo al menos uno de los dos y filtrando por `guild_id`
- [x] 4.6 Portar `desmarcar`: UPDATE de `vista_en` a NULL por nombre o por id, filtrado por `guild_id`
- [x] 4.7 Revisar que cada handler responde una sola vez (el código Go actual encadena varios `respond` en `desmarcar` sin cortar el flujo)

## 5. Autocompletado

- [x] 5.1 Portar el proveedor de sugerencias de películas pendientes (`vista_en IS NULL`) filtrado por `guild_id`
- [x] 5.2 Portar el proveedor de sugerencias de películas vistas (`vista_en IS NOT NULL`) filtrado por `guild_id`
- [x] 5.3 Unificar ambos en una función parametrizada por el estado de visto (resuelve el TODO heredado de `main.go`)
- [x] 5.4 Limitar las sugerencias a las 25 que admite Discord

## 6. Registro de comandos

- [x] 6.1 Escribir el script one-off que define los 6 comandos con sus opciones y flags de autocompletado
- [x] 6.2 Hacer que el script llame al endpoint REST de bulk overwrite de Discord con el token de bot
- [x] 6.3 Exponerlo como `npm run register` y documentar que solo se ejecuta al cambiar la definición de comandos

## 7. Despliegue y configuración en Discord

- [x] 7.1 Desplegar el worker con `wrangler deploy`. URL: `https://valstrax.raynou-dev.workers.dev`
- [x] 7.2 Verificar el handshake. La firma válida no se puede reproducir contra producción
      (la clave privada es de Discord), así que se comprobó en dos mitades: contra la URL
      desplegada, que `GET` da 405 y que un `POST` sin firma o con firma inválida da 401;
      y en local con un par Ed25519 desechable, que un `PING` bien firmado devuelve
      `200 {"type":1}` y que alterar el timestamp con la misma firma da 401
- [x] 7.3 Fijar el Interactions Endpoint URL en el Discord Developer Portal y confirmar que Discord lo acepta
- [x] 7.4 Ejecutar `npm run register` y comprobar que los comandos aparecen en Discord

## 8. Verificación end-to-end

- [x] 8.1 Probar los 6 comandos en un servidor real
- [x] 8.2 Probar los 3 autocompletados (`vista`, `quitar`, `desmarcar`)
- [x] 8.3 Comprobar el aislamiento entre dos servidores distintos
- [x] 8.4 Comprobar que una petición con firma manipulada recibe 401 (verificado contra la URL desplegada)
- [x] 8.5 Comprobar que un comando en DM responde el mensaje de error de guild ausente

## 9. Retirada del código Go

- [x] 9.1 Eliminar `main.go`, `go.mod`, `go.sum` y `scripts/setup_vs_debugger.go`
- [x] 9.2 Actualizar `README.md` con las instrucciones de despliegue del worker
- [x] 9.3 Actualizar `.gitignore` para el proyecto Node (`node_modules`, `.wrangler`) y retirar entradas del binario Go
- [x] 9.4 Actualizar `openspec/config.yaml` con el stack nuevo en el bloque `context`
