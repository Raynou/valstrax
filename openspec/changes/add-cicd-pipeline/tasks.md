## 1. Adoptar el sistema de migraciones de D1

- [x] 1.1 Crear la carpeta `migrations/` y mover `schema.sql` a `migrations/0001_crear_peliculas.sql`
      como copia exacta (mismos `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS`), para
      que la migración inicial sea un no-op idempotente sobre la base D1 remota, que ya tiene la
      tabla y el índice pero todavía no la tabla de control `d1_migrations`
- [x] 1.2 Declarar `migrations_dir = "migrations"` en `wrangler.toml`
- [x] 1.3 Sustituir los scripts `schema:local` / `schema:remote` de `package.json` por
      `migrate:local` (`wrangler d1 migrations apply valstrax-db --local`) y
      `migrate:remote` (`wrangler d1 migrations apply valstrax-db --remote`)
- [x] 1.4 Aplicar `migrate:local` sobre una base D1 local en blanco y comprobar que crea la tabla,
      el índice y la tabla de control `d1_migrations`
- [x] 1.5 Sobre la base D1 remota (esquema presente, `d1_migrations` ausente): ejecutar
      `migrate:remote` una vez y confirmar que wrangler crea `d1_migrations`, aplica `0001` sin
      error ni pérdida de datos y `migrations list --remote` la muestra como aplicada; una segunda
      ejecución no debe volver a aplicarla

## 2. Scripts npm reutilizables por la pipeline

- [x] 2.1 Verificar que `npm run typecheck`, `npm run migrate:remote` y `npm run deploy` funcionan
      de forma no interactiva tomando `CLOUDFLARE_API_TOKEN` del entorno
- [x] 2.2 Verificar que `npm run register` funciona en CI leyendo `DISCORD_TOKEN` y
      `DISCORD_APPLICATION_ID` del entorno (hoy `package.json` lo lanza con `--env-file=.env`;
      hacer que ese fichero sea opcional para que en local siga usando `.env` y en CI use el entorno)
- [x] 2.3 Añadir un script `ci` (o `release`) que encadene `typecheck` → `migrate:remote` →
      `deploy` → `register` para tener una entrada única reutilizable tanto por la Action como por
      un despliegue manual
- [x] 2.4 Comprobar que la ruta de despliegue manual de emergencia sigue documentada y operativa
      (los mismos scripts npm, sin depender de GitHub Actions)

## 3. Workflow de GitHub Actions

- [x] 3.1 Crear `.github/workflows/deploy-cloudflare.yml` disparado por `push` a `main`
- [x] 3.2 Pasos: `actions/checkout` → `actions/setup-node` (con caché de npm) → `npm ci`
- [x] 3.3 Paso de validación: `npm run typecheck` como gate previo al despliegue
- [x] 3.4 Paso de migraciones: `npm run migrate:remote` antes del despliegue, con
      `CLOUDFLARE_API_TOKEN` inyectado desde `secrets`
- [x] 3.5 Paso de despliegue: `npm run deploy`, con `CLOUDFLARE_API_TOKEN` desde `secrets`
- [x] 3.6 Paso de registro de comandos: `npm run register` después del deploy, con
      `DISCORD_TOKEN` y `DISCORD_APPLICATION_ID` desde `secrets` (sin `GUILD_ID`: registro global)
- [x] 3.7 Añadir `concurrency` para serializar despliegues y evitar carreras entre pushes seguidos
- [x] 3.8 Registrar en GitHub → Settings → Secrets and variables → Actions los tres secretos que
      necesita la pipeline: `CLOUDFLARE_API_TOKEN`, `DISCORD_TOKEN` y `DISCORD_APPLICATION_ID`
      (excepción a D3; el secreto de runtime del worker `DISCORD_PUBLIC_KEY` sigue solo en Cloudflare)

## 4. Documentación

- [x] 4.1 Actualizar `README.md`: reemplazar `npm run schema:*` por `npm run migrate:*` y describir
      que el despliegue a producción lo hace la Action en cada push a `main`
- [x] 4.2 Documentar en `README.md` la ruta de despliegue manual de emergencia (los mismos scripts
      npm, incluido `npm run register`, que la pipeline también ejecuta en cada despliegue)
- [x] 4.3 Documentar los tres secretos de la pipeline y cómo generarlos: `CLOUDFLARE_API_TOKEN`
      con permisos mínimos (Workers Scripts: Edit, D1: Edit), `DISCORD_TOKEN` y
      `DISCORD_APPLICATION_ID`; y dónde se configuran (GitHub Actions Secrets)

## 5. Verificación end-to-end

- [ ] 5.1 Hacer un push a `main` con un cambio trivial y comprobar que la Action despliega el worker
- [ ] 5.2 Añadir una migración `0002` de prueba (columna nueva o índice) y comprobar que la Action
      la aplica antes del deploy y que `migrations list --remote` la marca como aplicada
- [ ] 5.3 Forzar un fallo de `typecheck` en una rama y comprobar que, al fusionar, la Action corta
      antes de tocar la base de datos o desplegar
- [ ] 5.4 Comprobar que un segundo push mientras corre un despliegue queda en cola por `concurrency`
      y no se solapa
- [ ] 5.5 Modificar la definición de un slash command, hacer push a `main` y comprobar que la Action
      ejecuta `register` y que Discord refleja el cambio
