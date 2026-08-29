# http-interactions Specification

## Purpose
TBD - created by archiving change switch-to-serverless. Update Purpose after archive.
## Requirements
### Requirement: Verificación de firma de las peticiones entrantes

El worker SHALL verificar la firma Ed25519 de toda petición entrante usando la clave pública de la aplicación de Discord, antes de interpretar su contenido. Las peticiones cuya firma sea inválida, esté ausente, o carezca de timestamp DEBEN ser rechazadas con HTTP 401 sin ejecutar ninguna lógica de negocio ni consulta a la base de datos.

La verificación SHALL realizarse sobre el cuerpo crudo de la petición concatenado al timestamp, leído una única vez y sin parsear previamente.

#### Scenario: Petición con firma válida

- **WHEN** llega una petición con las cabeceras `X-Signature-Ed25519` y `X-Signature-Timestamp` cuya firma corresponde al cuerpo enviado
- **THEN** el worker procesa la interacción y responde con el resultado del comando

#### Scenario: Petición con firma inválida

- **WHEN** llega una petición cuya cabecera `X-Signature-Ed25519` no corresponde al cuerpo enviado
- **THEN** el worker responde HTTP 401 y no ejecuta ninguna consulta contra la base de datos

#### Scenario: Petición sin cabeceras de firma

- **WHEN** llega una petición sin `X-Signature-Ed25519` o sin `X-Signature-Timestamp`
- **THEN** el worker responde HTTP 401 sin intentar parsear el cuerpo

#### Scenario: El cuerpo se verifica antes de parsearse

- **WHEN** el worker recibe cualquier petición
- **THEN** la verificación de firma opera sobre el texto crudo del cuerpo, no sobre una reserialización del JSON parseado

### Requirement: Handshake de validación del endpoint

El worker SHALL responder al tipo de interacción `PING` (tipo 1) con una respuesta `PONG` (tipo 1), de modo que Discord pueda validar y aceptar el Interactions Endpoint URL.

#### Scenario: Discord valida la URL del endpoint

- **WHEN** Discord envía una interacción de tipo `PING` correctamente firmada
- **THEN** el worker responde HTTP 200 con el cuerpo JSON `{"type": 1}`

#### Scenario: PING con firma inválida

- **WHEN** Discord envía una interacción de tipo `PING` con firma inválida
- **THEN** el worker responde HTTP 401 y Discord rechaza la configuración del endpoint

### Requirement: Enrutado de comandos y autocompletado

El worker SHALL enrutar cada interacción según su tipo: las de tipo comando de aplicación al handler correspondiente a su nombre, y las de tipo autocompletado al proveedor de sugerencias correspondiente. Una interacción cuyo nombre de comando no tenga handler registrado NO DEBE provocar un error no controlado.

#### Scenario: Comando conocido

- **WHEN** llega una interacción de comando con nombre `agregar`, `sugerir`, `vista`, `lista`, `quitar` o `desmarcar`
- **THEN** el worker ejecuta el handler correspondiente y devuelve su respuesta

#### Scenario: Interacción de autocompletado

- **WHEN** llega una interacción de tipo autocompletado para `vista`, `quitar` o `desmarcar`
- **THEN** el worker responde con una lista de sugerencias, sin ejecutar el handler del comando

#### Scenario: Comando desconocido

- **WHEN** llega una interacción de comando cuyo nombre no corresponde a ningún handler
- **THEN** el worker responde con un mensaje de error controlado en lugar de fallar

### Requirement: Respuesta dentro de la ventana de Discord

El worker SHALL emitir su respuesta inicial a toda interacción en menos de 3 segundos, respondiendo de forma síncrona sin recurrir a respuestas diferidas.

#### Scenario: Respuesta síncrona a un comando

- **WHEN** un usuario ejecuta cualquiera de los slash commands
- **THEN** el worker devuelve una respuesta de tipo `CHANNEL_MESSAGE_WITH_SOURCE` en la misma respuesta HTTP, sin emitir un webhook de seguimiento

