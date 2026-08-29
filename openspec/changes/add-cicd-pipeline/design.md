## Context

Actualmente, el proyecto se despliega manualmente haciendo uso de los comandos de `wrangler` y los comandos custom para propagar a discord los nuevos comandos que se vayan añadiendo/modificando al bot; además de que el esquema debe de ser actualizado manualmente. Este proceso, además de ser tedioso, es propenso a errores y olvidos, por lo que se busca automatizar este proceso de despliegue con un flujo basado en GitHub Actions.

## Goals / Non-Goals

**Goals**
- Eliminar el proceso manual para hacer despliegue a cloudfare, agilizando la entrega de cambios por muy complejos que sean.
- Preservar la superficie de comandos actual existente en el proyecto para dejar la puerta abierta a hacer un despliegue manual de emergencia, estos comandos deben ser reaprovechados por la linea ci/cd.

**Non-Goals**
- Añadir pruebas unitarias al proyecto y agregarlas como un gatekeep a la linea CI/CD, eso se puede discutir como un cambio aparte.
- Dockerizar el bot, puesto que es innecesario ya que utiliza arquitectura serverless.
- Generación de releases o release notes, puesto que es un proyecto muy pequeño todavia y no lo amerita aún. Se deja la puerta abierta en un futuro para implementar esto.
- Generación de un artifact, ya que no aporta valor tener el compilado en GitHub.

## Decisions

### D1: Usage of main branch for deployment

Los flujos de trabajo de GitHub Actions serán ejecutados en cada `push` a la rama `main`, favoreciendo un flujo de desarrollo tipo `trunk` o `gitflow`.

### D2: Migrations with Wrangler (no ORM)

Implementar el uso de un ORM que soporte un sistema de migraciones (TypeORM, Sequelize, Prism, etc) es un overhead innecesario que implicaria reescribir parte de la lógica de acceso a la capa de datos que ya tenemos; por el contrario, el uso de `wrangler` para esto simplifica el proceso de migraciones sin tener que hacer mayores cambios a nuestra base de código.

Documentación al respecto: https://developers.cloudflare.com/d1/reference/migrations/

**Adopción sobre la base ya provisionada.** Hoy la D1 remota ya tiene la tabla `peliculas` y su
índice (aplicados a mano con `wrangler d1 execute --file=schema.sql`) pero **no existe la tabla
`d1_migrations`**. La estrategia es que `0001_crear_peliculas.sql` sea una copia exacta del
`schema.sql` actual, con sus `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS`
intactos. En el primer `wrangler d1 migrations apply --remote`, wrangler crea `d1_migrations`,
ejecuta `0001` (que es un no-op sobre el esquema existente) y la marca como aplicada. No hace
falta ningún baseline manual. Esto es seguro porque la tabla viva se creó de ese mismo fichero,
así que no hay *drift* que `IF NOT EXISTS` pueda ocultar.

### D3: Secret management

El secreto que el worker lee en runtime (`DISCORD_PUBLIC_KEY`) se gestiona directamente en las configuraciones de Cloudfare (`wrangler secret put`), no en GitHub, por simplicidad.

Como excepción, las credenciales que la propia pipeline necesita para operar sí viven en GitHub Secrets, porque no hay otra forma de que la Action se autentique:

- `CLOUDFLARE_API_TOKEN`: lo usa `wrangler` para aplicar migraciones y desplegar. Permisos mínimos: *Workers Scripts: Edit* y *D1: Edit*.
- `DISCORD_TOKEN` y `DISCORD_APPLICATION_ID`: los usa `scripts/register.ts` para registrar los slash commands desde la pipeline.

### D4: Command registration in the pipeline

`npm run register` se ejecuta en cada despliegue. La llamada a Discord es un *bulk overwrite* idempotente: reenviar la misma definición de comandos no tiene efecto, así que no hace falta detectar cambios en `scripts/register.ts`. El registro es global (sin `GUILD_ID`) y corre después del `deploy`, al ser independiente del código del worker.

## Risks / Trade-offs

- **Se promueve más acoplamiento de Cloudfare** → Al tener tanto acoplamiento, no solo a nivel de código sino a nivel de infraestructura, se entiende que en un futuro si se requiere migrar de proveedor se tendrán que hacer cambios significativos, no solo a la pipeline, sino que a la arquitectura del proyecto. Al ser un proyecto pequeño y en fase temprana, se es consciente de esto y se acepta.

## Migration plan

1. Modificar la estructura del proyecto y el nombre del fichero `schema.sql` para que vaya acorde al ejemplo mostrado en la documentación que se fue proporcionada en D2
2. Crear los comandos necesarios de NPM para facilitarle la vida a la linea CI/CD
3. Crera la carpeta `.github` con todos los archivos necesarios