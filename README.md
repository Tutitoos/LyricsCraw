# LyricsCraw (SpotyCraw)

API REST en Go para obtener letras de canciones mediante scraping de Vagalume usando Chrome (chromedp). Incluye cache en memoria con TTL para reducir latencia en consultas repetidas.

## Características

- Gin HTTP server con endpoints `/health` y `/v1/lyrics`.
- Scraping con `chromedp` (Chrome en modo visible por defecto).
- Cache en memoria con TTL y limpiador en background.
- Logging estructurado con Uber Zap.

## Requisitos

- Go (el módulo declara `go 1.24.4`).
- Google Chrome o Chromium instalado en el sistema (chromedp lo usa bajo el capó).
- macOS, Linux o Windows.

## Variables de entorno

- `APP_ENV`: `development` o `production`. Afecta logs y modo Gin. Por defecto: `development`.
- `APP_PORT`: puerto del servidor. Por defecto: `8080`.
- `APP_LYRICS_CACHE_TTL_SECONDS`: TTL del cache en segundos. Por defecto: `1800` (30 min).
- `APP_LYRICS_CACHE_MAX_ENTRIES`: capacidad máxima del cache. Por defecto: `1000`.

Puedes definirlas en un `.env` en la raíz; si existe, se carga al iniciar.

## Instalar y ejecutar

Clona el repositorio y descarga dependencias.

Desarrollo (con recarga manual):

```bash
make dev
```

Producción (binario en `./bin/app`):

```bash
make build
make start
```

Alternativa sin Makefile:

```bash
# development
sh scripts/dev.sh

# production
sh scripts/build.sh
sh scripts/start.sh
```

## Endpoints

Health check:

```http
GET /health -> 200 { "status": "ok" }
```

Obtener letras:

```http
GET /v1/lyrics?query=<artista> - <cancion>
```

Ejemplo curl:

```bash
curl "http://localhost:8080/v1/lyrics?query=Coldplay%20-%20Yellow"
```

Respuesta:

```json
{
	"data": "<letra>",
	"cached": true
}
```

La clave del cache es la `query` normalizada (minúsculas, trim). Si hay hit en cache, `cached` es `true` y la respuesta es inmediata; de lo contrario se realiza scraping y se guarda.

## Cómo funciona el scraping (resumen)

1) Abre un contexto de Chrome (`headless=false` por defecto) con un User-Agent aleatorio.
2) Navega a `https://www.vagalume.com.br/search?q=<query>` y obtiene el primer resultado.
3) Ajusta la URL (elimina `-traducao` si aplica) y navega al detalle.
4) Si hay aviso +18, intenta aceptar el modal.
5) Extrae el texto de `div#lyrics` y limpia mensajes de confirmación.

Código relevante:
- `src/scraper/Scraper.go`
- `src/scraper/UserAgentGenerator.go`

## Notas sobre Chrome

- Actualmente cada solicitud crea un contexto nuevo de Chrome y se cierra al terminar. Esto es seguro y simple.
- Si necesitas mantener Chrome abierto persistentemente para varias solicitudes, considera crear un contexto global reutilizable y controlar su ciclo de vida. (No implementado por defecto.)

## Cache en memoria

- Implementación en `src/cache/LyricsCache.go`.
- Inicialización automática en `main` leyendo variables de entorno.
- Limpieza periódica en background (~TTL/2, mínimo 30s).

## Estructura del proyecto

- `src/main.go`: arranque del servidor, carga `.env`, logger y router.
- `src/api/router/Router.go`: rutas y grupos.
- `src/api/controller/TokenController.go`: controlador de letras.
- `src/scraper/*`: scraping y user-agent.
- `src/logger/logger.go`: configuración de Zap.
- `scripts/*`: helpers para dev/build/start.
- `bruno-http/*`: colección Bruno opcional para probar.

## Problemas comunes y solución

- Chrome no encontrado o arranque lento: instala Google Chrome estable y mantenlo actualizado. En contenedores, habilita flags como `--no-sandbox` (ya configurado en el código) según sea necesario.
- Permisos en macOS: si aparece diálogo de seguridad al lanzar Chrome, autoriza la app.
- Selectores rotos: los selectores de Vagalume pueden cambiar. Revisa `a.gs-title`, `div#lyrics` y el modal +18 si deja de funcionar.

## Licencia

No especificada.
