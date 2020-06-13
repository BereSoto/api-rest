# Prueba

Instalar dependencias

```bash
go get "github.com/buaazp/fasthttprouter"
go get "github.com/valyala/fasthttp"
go get "github.com/lib/pq"
go get -u "github.com/likexian/whois-go"
```

Inizializar cockroach

```bash
cockroach start-single-node \
    --insecure \
    --listen-addr=localhost:26257 \
    --http-addr=localhost:8080 \
    --background
```

Inizializar usuarios y bases de datos

```bash
echo "CREATE USER IF NOT EXISTS test; CREATE DATABASE domains; GRANT ALL ON DATABASE domains TO test;" | cockroach sql --insecure
```

Arrancar server:

```bash
go run main.go
```
