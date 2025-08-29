# Configuration Guide

This application supports flexible configuration through multiple sources with the following priority order:

1. **Environment Variables** (highest priority)
2. **Configuration Files** (YAML)
3. **Default Values** (lowest priority)

## Environment Variables

The configuration uses `setEnvKeyReplacer` to convert nested config keys to environment variable format:

- Config key: `restServer.port` → Environment variable: `REST_SERVER_PORT`
- Config key: `postgres.host` → Environment variable: `POSTGRES_HOST`

### Why No Prefix?

We **don't use a prefix** like `ONEFEED_` because:
- Keeps environment variable names shorter and cleaner
- Reduces typing and potential errors
- Makes variables more readable in deployment configs
- Follows the principle of clear, self-documenting names

### Available Environment Variables

#### Server Configuration
```bash
REST_SERVER_PORT=8080           # HTTP server port
```

#### PostgreSQL Configuration
```bash
POSTGRES_HOST=localhost         # Database host
POSTGRES_PORT=5432             # Database port
POSTGRES_USER=postgres         # Database username (REQUIRED - no default)
POSTGRES_PASSWORD=secret       # Database password (REQUIRED - no default)
POSTGRES_DBNAME=onefeed        # Database name (REQUIRED - no default)

# PostgreSQL Connection Pool Settings (optional - have sensible defaults)
POSTGRES_POOL_MAX_CONNS=25              # Maximum connections
POSTGRES_POOL_MIN_CONNS=5               # Minimum connections
POSTGRES_POOL_MAX_CONN_LIFETIME=60      # Max connection lifetime (minutes)
POSTGRES_POOL_MAX_CONN_IDLE_TIME=30     # Max idle time (minutes)
POSTGRES_POOL_HEALTH_CHECK_PERIOD=1     # Health check interval (minutes)
POSTGRES_POOL_CONNECT_TIMEOUT=5         # Connection timeout (seconds)
```

#### Redis Configuration
```bash
REDIS_HOST=localhost           # Redis host
REDIS_PORT=6379               # Redis port
REDIS_PASSWORD=secret         # Redis password (optional if no auth)

# Redis Connection Pool Settings (optional - have sensible defaults)
REDIS_POOL_POOL_SIZE=15                 # Maximum socket connections
REDIS_POOL_MIN_IDLE_CONNS=5             # Minimum idle connections
REDIS_POOL_MAX_IDLE_CONNS=10            # Maximum idle connections
REDIS_POOL_POOL_TIMEOUT=4               # Pool timeout (seconds)
REDIS_POOL_IDLE_TIMEOUT=5               # Idle timeout (minutes)
REDIS_POOL_MAX_CONN_AGE=30              # Max connection age (minutes)
REDIS_POOL_DIAL_TIMEOUT=5               # Dial timeout (seconds)
REDIS_POOL_READ_TIMEOUT=3               # Read timeout (seconds)
REDIS_POOL_WRITE_TIMEOUT=3              # Write timeout (seconds)
REDIS_POOL_MAX_RETRIES=2                # Maximum retries
REDIS_POOL_MIN_RETRY_BACKOFF=8          # Min retry backoff (milliseconds)
REDIS_POOL_MAX_RETRY_BACKOFF=512        # Max retry backoff (milliseconds)
```

## Configuration File (config.yaml)

```yaml
restServer:
  port: 8080

postgres:
  host: localhost
  port: 5432
  user: postgres      # REQUIRED - no default
  password: secret    # REQUIRED - no default
  dbname: onefeed     # REQUIRED - no default
  pool:               # Optional - sensible defaults provided
    maxConns: 25
    minConns: 5
    maxConnLifetime: 60      # minutes
    maxConnIdleTime: 30      # minutes
    healthCheckPeriod: 1     # minutes
    connectTimeout: 5        # seconds

redis:
  host: localhost
  port: 6379
  password: secret    # Optional - only if Redis requires auth
  pool:               # Optional - sensible defaults provided
    poolSize: 15
    minIdleConns: 5
    maxIdleConns: 10
    poolTimeout: 4           # seconds
    idleTimeout: 5           # minutes
    maxConnAge: 30           # minutes
    dialTimeout: 5           # seconds
    readTimeout: 3           # seconds
    writeTimeout: 3          # seconds
    maxRetries: 2
    minRetryBackoff: 8       # milliseconds
    maxRetryBackoff: 512     # milliseconds
```

## Docker/Container Deployment

For containerized deployments, you can use environment variables only:

```bash
docker run -e POSTGRES_HOST=db \
           -e POSTGRES_PASSWORD=mypassword \
           -e REDIS_HOST=redis \
           onefeed-app
```

## Default Values

The application provides sensible defaults for development:

- Server Port: `8080`
- PostgreSQL Host: `localhost:5432` 
- Redis Host: `localhost:6379`

**Security Note**: Credentials (username, password, database name) have **NO defaults** and must be explicitly provided via environment variables or config file.

**Pool Defaults**: Both PostgreSQL and Redis connection pools have optimized defaults suitable for production use.

## Benefits of This Approach

1. **12-Factor App Compliance**: Configuration through environment variables
2. **Development Friendly**: Sensible defaults for local development
3. **Production Ready**: Easy to override with environment variables
4. **Container Ready**: Works seamlessly with Docker/Kubernetes
5. **Flexible**: Supports both file-based and env-based configuration