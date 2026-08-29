## Why

Actualmente, el proyecto se despliega manualmente utilizando los comandos de `wrangler`, esto es un proceso manual y propenso a equivaciones, por lo que añadir una CI/CD pipeline basada en `GitHub Actions` y con las propias herramientas que brinda `wrangler` es un buena idea.

## What changes

- Se añade nuevo flujo de trabajo para el build, la actualizacion del esquema de base de datos (si es necesario), el despliegue del bot a cloudflare y la propagación de los slash commands a Discord.

## Capabilites

### New Capabilities

- `ci-cd-pipeline`: Ejecuta migraciones a la base de datos, buildea a una nueva versión del proyecto, despliega el compilado en Cloudfare workers y registra los slash commands en Discord.

### Modified Capabilites

Ninguna

## Impact

- **Nueva carpeta**: Se crea una carpeta `.github` con el workfow para desplegar a cloudfare (`deploy-cloudfare.yml`)
- **Datos**: Se adopta el sistema de migraciones proporcionado por `wrangler` para manejar y versionar los cambios en nuestro esquema, tal y como lo muestra la documentación de cloudfare: https://developers.cloudflare.com/d1/reference/migrations/