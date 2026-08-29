## ADDED Requirements

### Requirement: Despliegue automático en push a main

El sistema SHALL desplegar el worker a Cloudflare de forma automática en cada `push` a la rama
`main`, sin intervención manual. Ningún otro evento (push a otras ramas, pull request, tag)
dispara un despliegue.

#### Scenario: Un push a main despliega el worker

- **WHEN** se integra un commit en `main`
- **THEN** GitHub Actions ejecuta el flujo de despliegue y publica la nueva versión del worker
  en Cloudflare

#### Scenario: Un push a una rama de trabajo no despliega

- **WHEN** se hace `push` a una rama distinta de `main`
- **THEN** el flujo de despliegue no se ejecuta y no se toca ni la base de datos ni el worker en
  producción

#### Scenario: Despliegues serializados

- **WHEN** llegan dos pushes a `main` con poca diferencia de tiempo
- **THEN** el segundo despliegue espera a que termine el primero en lugar de solaparse

### Requirement: Migraciones de esquema antes del despliegue

El flujo SHALL aplicar las migraciones pendientes de la base D1 remota antes de publicar el
worker. El esquema SHALL versionarse con el sistema de migraciones de `wrangler` (carpeta
`migrations/` con ficheros numerados), no con un único fichero `schema.sql` reejecutable.

#### Scenario: Una migración pendiente se aplica antes del deploy

- **WHEN** el commit integrado en `main` añade un fichero de migración nuevo en `migrations/`
- **THEN** el flujo ejecuta `wrangler d1 migrations apply` contra la base remota y solo después
  publica el worker

#### Scenario: Sin migraciones pendientes

- **WHEN** el commit integrado no añade migraciones
- **THEN** el paso de migraciones no aplica cambios y el flujo continúa hasta el despliegue

#### Scenario: Fallo de migración detiene el despliegue

- **WHEN** una migración falla al aplicarse
- **THEN** el flujo se detiene con error y no se publica la nueva versión del worker

#### Scenario: Adopción de migraciones sobre la base ya provisionada

- **WHEN** se ejecuta `migrate:remote` por primera vez sobre la base D1 remota, que ya contiene
  la tabla `peliculas` y su índice del despliegue manual previo pero todavía no la tabla de
  control `d1_migrations`
- **THEN** `wrangler` crea `d1_migrations`, ejecuta `0001` como no-op (sin error ni pérdida de
  datos) y la registra como aplicada

#### Scenario: La migración inicial no se reejecuta

- **WHEN** se ejecuta `migrate:remote` una segunda vez sin añadir migraciones nuevas
- **THEN** `0001` no se vuelve a aplicar porque ya consta en `d1_migrations`

### Requirement: Validación previa como gate del despliegue

El flujo SHALL ejecutar la comprobación de tipos (`npm run typecheck`) antes de aplicar
migraciones o desplegar, y SHALL abortar sin efectos secundarios si esa comprobación falla.

#### Scenario: Un error de tipos corta el flujo

- **WHEN** el código integrado en `main` no compila con `tsc --noEmit`
- **THEN** el flujo termina en error antes de tocar la base de datos remota o publicar el worker

### Requirement: Propagación de slash commands en cada despliegue

El flujo SHALL registrar los slash commands en Discord (`npm run register`) después de publicar
el worker. El registro es global y la llamada de *bulk overwrite* a Discord SHALL ser
idempotente: reejecutarla con la misma definición de comandos no produce cambios.

#### Scenario: Un cambio en la definición de comandos se propaga

- **WHEN** el commit integrado en `main` modifica la definición de los slash commands en
  `scripts/register.ts`
- **THEN** el flujo ejecuta `npm run register` tras el despliegue y Discord queda con la nueva
  definición

#### Scenario: Despliegue sin cambios en los comandos

- **WHEN** el commit integrado no toca la definición de los slash commands
- **THEN** el paso de registro se ejecuta igualmente y no altera los comandos ya registrados

### Requirement: Reutilización de la superficie de comandos para despliegue manual

Los pasos del flujo SHALL invocarse a través de los scripts de `package.json`
(`typecheck`, `migrate:remote`, `deploy`, `register`), de modo que un operador pueda reproducir
un despliegue de emergencia ejecutando los mismos comandos en local. El flujo NO DEBE contener
lógica de despliegue que no exista también como script npm.

#### Scenario: Despliegue manual de emergencia

- **WHEN** GitHub Actions no está disponible y hace falta desplegar
- **THEN** un operador con las credenciales necesarias puede ejecutar `npm run migrate:remote`,
  `npm run deploy` y `npm run register` en local y obtener el mismo resultado que el flujo
  automático

### Requirement: Gestión de credenciales

Los secretos que necesita la pipeline SHALL almacenarse en GitHub Actions Secrets:
`CLOUDFLARE_API_TOKEN` (permisos *Workers Scripts: Edit* y *D1: Edit*), `DISCORD_TOKEN` y
`DISCORD_APPLICATION_ID`. El único secreto que el worker lee en runtime, `DISCORD_PUBLIC_KEY`,
SHALL gestionarse en Cloudflare (`wrangler secret put`) y NO DEBE aparecer en GitHub.

#### Scenario: El flujo se autentica con el token de Cloudflare

- **WHEN** el flujo ejecuta cualquier comando de `wrangler` contra Cloudflare
- **THEN** `wrangler` toma `CLOUDFLARE_API_TOKEN` del entorno inyectado desde GitHub Secrets

#### Scenario: El registro de comandos usa los secretos de Discord

- **WHEN** el flujo ejecuta `npm run register`
- **THEN** `scripts/register.ts` toma `DISCORD_TOKEN` y `DISCORD_APPLICATION_ID` del entorno
  inyectado desde GitHub Secrets

#### Scenario: El secreto de runtime del worker no pasa por GitHub

- **WHEN** se revisa la configuración de GitHub Secrets del repositorio
- **THEN** no aparece `DISCORD_PUBLIC_KEY`; ese secreto vive solo en Cloudflare
