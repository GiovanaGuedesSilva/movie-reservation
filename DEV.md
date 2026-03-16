# Guia de Desenvolvimento — Movie Reservation System

Este documento detalha as etapas de desenvolvimento do projeto para que **Dev 1** e **Dev 2** possam trabalhar em paralelo com clareza sobre responsabilidades, dependências e critérios de conclusão.

---

## Divisão de Domínios

| | Dev 1 | Dev 2 |
|---|---|---|
| **Domínio principal** | Auth · Usuários · Filmes · Gêneros | Salas · Assentos · Sessões · Reservas · Relatórios |
| **Camadas** | domain → ports → service → adapter | domain → ports → service → adapter |

> **Regra de ouro:** nenhuma etapa de negócio começa sem a **Etapa 0** estar concluída por ambas.

---

## Visão Geral das Etapas

```
Etapa 0 ──────────── Setup do projeto (AMBAS, ~1 dia)
    │
    ├── Etapa 1 Dev1: Domain User/Movie/Genre + Ports
    ├── Etapa 1 Dev2: Domain Theater/Seat/Showtime/Reservation + Ports
    │
    ├── Etapa 2 Dev1: Services Auth/User/Movie + Repos GORM
    ├── Etapa 2 Dev2: Services Theater/Showtime + Repos GORM
    │
    ├── Etapa 3 Dev1: HTTP Handlers Auth/User/Movie + Middleware
    ├── Etapa 3 Dev2: HTTP Handlers Theater/Showtime + Router
    │
    ├── Etapa 4 Dev1: Testes de Auth e Movie
    ├── Etapa 4 Dev2: Service/Handler de Reservation (concorrência)
    │
    ├── Etapa 5 Dev1: Relatórios de receita + seed de produção
    ├── Etapa 5 Dev2: Relatórios de capacidade + Postman Collection
    │
    └── Etapa 6 ──── Integração final, revisão, Docker prod (AMBAS)
```

---

## Etapa 0 — Setup do Projeto

> **Responsáveis:** Ambas  
> **Estimativa:** 1 dia  
> **Branch:** `develop` (trabalho direto)

Esta etapa deve ser feita em conjunto ou com uma fazendo e a outra revisando via PR.

### Checklist

- [ ] Criar repositório no GitHub com branch `main` e `develop`
- [ ] Definir regras de branch protection em `main` (exige PR + aprovação)
- [ ] Inicializar módulo Go

```bash
mkdir movie-reservation && cd movie-reservation
go mod init github.com/seu-usuario/movie-reservation
```

- [ ] Criar estrutura completa de pastas (veja README.md)
- [ ] Configurar `docker-compose.yml` com PostgreSQL 15
- [ ] Criar `Dockerfile` multi-stage para a aplicação Go
- [ ] Criar `.env.example` com todas as variáveis necessárias
- [ ] Criar `Makefile` com todos os comandos do projeto
- [ ] Instalar dependências base

```bash
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get github.com/joho/godotenv
go get github.com/google/uuid
go get golang.org/x/crypto/bcrypt
go get github.com/golang-migrate/migrate/v4
go get github.com/stretchr/testify
```

- [ ] Criar `pkg/config/config.go` — leitura de `.env`
- [ ] Criar `pkg/logger/logger.go` — logger estruturado básico
- [ ] Criar `pkg/apperrors/errors.go` — erros de domínio tipados
- [ ] Criar `cmd/api/main.go` com servidor HTTP mínimo rodando

### Critério de conclusão

`make docker-up` sobe o banco e a API responde `200` em `GET /health`.

---

## Etapa 1 — Entidades de Domínio e Ports

> **Estimativa:** 1–2 dias  
> **Pré-requisito:** Etapa 0 concluída

### Dev 1 — Branch: `feature/domain-auth-movies`

**Criar as entidades de domínio:**

- [ ] `internal/core/domain/user.go`
  ```go
  type Role string
  const (RoleAdmin Role = "admin"; RoleUser Role = "user")

  type User struct {
      ID           uuid.UUID
      Name         string
      Email        string
      PasswordHash string
      Role         Role
      CreatedAt    time.Time
      UpdatedAt    time.Time
  }
  ```

- [ ] `internal/core/domain/movie.go`
  ```go
  type Movie struct {
      ID              uuid.UUID
      Title           string
      Description     string
      PosterURL       string
      DurationMinutes int
      Genres          []Genre
      CreatedAt       time.Time
      UpdatedAt       time.Time
  }
  ```

- [ ] `internal/core/domain/genre.go`
  ```go
  type Genre struct {
      ID   uuid.UUID
      Name string
  }
  ```

**Criar as interfaces (ports):**

- [ ] `internal/core/ports/inbound/auth_service.go`
  ```go
  type AuthService interface {
      SignUp(ctx context.Context, input SignUpInput) (*domain.User, error)
      Login(ctx context.Context, input LoginInput) (string, error) // retorna JWT
  }
  ```

- [ ] `internal/core/ports/inbound/user_service.go`
  ```go
  type UserService interface {
      GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
      ListAll(ctx context.Context) ([]domain.User, error)
      PromoteToAdmin(ctx context.Context, id uuid.UUID) error
  }
  ```

- [ ] `internal/core/ports/inbound/movie_service.go`
  ```go
  type MovieService interface {
      Create(ctx context.Context, input CreateMovieInput) (*domain.Movie, error)
      Update(ctx context.Context, id uuid.UUID, input UpdateMovieInput) (*domain.Movie, error)
      Delete(ctx context.Context, id uuid.UUID) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Movie, error)
      List(ctx context.Context, filters MovieFilters) ([]domain.Movie, error)
  }
  ```

- [ ] `internal/core/ports/outbound/user_repository.go`
- [ ] `internal/core/ports/outbound/movie_repository.go`

### Dev 2 — Branch: `feature/domain-showtime-reservation`

- [ ] `internal/core/domain/theater.go`
  ```go
  type Theater struct {
      ID          uuid.UUID
      Name        string
      TotalRows   int
      SeatsPerRow int
      Seats       []Seat
  }
  ```

- [ ] `internal/core/domain/seat.go`
  ```go
  type Seat struct {
      ID        uuid.UUID
      TheaterID uuid.UUID
      Row       string  // "A", "B", ...
      Number    int     // 1, 2, ...
  }
  ```

- [ ] `internal/core/domain/showtime.go`
  ```go
  type Showtime struct {
      ID        uuid.UUID
      MovieID   uuid.UUID
      TheaterID uuid.UUID
      StartTime time.Time
      EndTime   time.Time
      Price     float64
  }
  ```

- [ ] `internal/core/domain/reservation.go`
  ```go
  type ReservationStatus string
  const (StatusActive ReservationStatus = "active"; StatusCancelled = "cancelled")

  type Reservation struct {
      ID         uuid.UUID
      UserID     uuid.UUID
      ShowtimeID uuid.UUID
      Seats      []Seat
      TotalPrice float64
      Status     ReservationStatus
      CreatedAt  time.Time
  }
  ```

- [ ] `internal/core/ports/inbound/showtime_service.go`
- [ ] `internal/core/ports/inbound/reservation_service.go`
- [ ] `internal/core/ports/outbound/theater_repository.go`
- [ ] `internal/core/ports/outbound/showtime_repository.go`
- [ ] `internal/core/ports/outbound/reservation_repository.go`

### Critério de conclusão

Todas as interfaces compilam sem erros. `go build ./...` passa.

---

## Etapa 2 — Services e Repositórios GORM

> **Estimativa:** 2–3 dias  
> **Pré-requisito:** Etapa 1 (pode começar com as interfaces já definidas, antes de merges)

### Dev 1 — Branch: `feature/service-auth-movies`

**Migrations:**

- [ ] `migrations/000001_create_users.up.sql`
- [ ] `migrations/000002_create_movies.up.sql`
- [ ] `migrations/000003_create_genres.up.sql`
- [ ] `migrations/000004_create_movie_genres.up.sql`

**Repositórios GORM:**

- [ ] `internal/adapters/secondary/postgres/models.go` — modelos GORM (UserModel, MovieModel...)
- [ ] `internal/adapters/secondary/postgres/db.go` — conexão GORM + AutoMigrate
- [ ] `internal/adapters/secondary/postgres/user_repository.go`
  - `Create`, `GetByID`, `GetByEmail`, `List`, `Update`
- [ ] `internal/adapters/secondary/postgres/movie_repository.go`
  - `Create`, `GetByID`, `List` (com filtro por gênero), `Update`, `Delete`

**Services:**

- [ ] `pkg/auth/jwt.go` — `GenerateToken(user)`, `ValidateToken(token)`
- [ ] `internal/core/services/auth_service.go`
  - `SignUp`: valida email único, faz hash da senha com bcrypt, salva usuário
  - `Login`: busca por email, compara hash, gera JWT
- [ ] `internal/core/services/user_service.go`
  - `GetByID`, `ListAll`, `PromoteToAdmin`
- [ ] `internal/core/services/movie_service.go`
  - CRUD completo, associa gêneros

### Dev 2 — Branch: `feature/service-showtime-theater`

**Migrations:**

- [ ] `migrations/000005_create_theaters.up.sql`
- [ ] `migrations/000006_create_seats.up.sql`
- [ ] `migrations/000007_create_showtimes.up.sql`
- [ ] `migrations/000008_create_reservations.up.sql`
- [ ] `migrations/000009_create_reservation_seats.up.sql`
  - Incluir: `UNIQUE(seat_id, showtime_id)` — chave para prevenção de overbooking

**Repositórios GORM:**

- [ ] `internal/adapters/secondary/postgres/theater_repository.go`
  - `Create` (cria sala E gera todos os assentos automaticamente), `GetByID`, `List`
- [ ] `internal/adapters/secondary/postgres/showtime_repository.go`
  - `Create`, `GetByID`, `ListByDate`, `ListByMovie`, `Update`, `Delete`
  - `GetAvailableSeats(showtimeID)` — retorna assentos não reservados
- [ ] `internal/adapters/secondary/postgres/reservation_repository.go`
  - `Create` (dentro de transação), `GetByID`, `ListByUser`, `ListAll`, `Cancel`
  - `IsSeatsAvailable(showtimeID, seatIDs)` — verifica disponibilidade com `SELECT FOR UPDATE`

**Services:**

- [ ] `internal/core/services/theater_service.go`
- [ ] `internal/core/services/showtime_service.go`
  - Valida conflito de horário na mesma sala
  - Calcula `end_time` baseado em `start_time + movie.duration_minutes`

### Critério de conclusão

`make migrate-up && make seed` executa sem erros. Testes unitários dos services passam com mocks.

---

## Etapa 3 — HTTP Handlers e Roteamento

> **Estimativa:** 2 dias  
> **Pré-requisito:** Etapa 2

### Dev 1 — Branch: `feature/handlers-auth-movies`

- [ ] `internal/adapters/primary/http/middleware/auth.go`
  - Extrai `Authorization: Bearer <token>`, valida JWT, injeta user no contexto
- [ ] `internal/adapters/primary/http/middleware/role.go`
  - Middleware `RequireAdmin` — retorna 403 se role != "admin"
- [ ] `internal/adapters/primary/http/handlers/auth_handler.go`
  - `POST /auth/signup`
  - `POST /auth/login`
- [ ] `internal/adapters/primary/http/handlers/user_handler.go`
  - `GET /users/me`
  - `GET /users` (admin)
  - `PUT /users/:id/promote` (admin)
- [ ] `internal/adapters/primary/http/handlers/movie_handler.go`
  - `GET /movies`, `GET /movies/:id`
  - `POST /movies`, `PUT /movies/:id`, `DELETE /movies/:id` (admin)
  - `GET /genres`, `POST /genres` (admin)

### Dev 2 — Branch: `feature/handlers-showtime-theater`

- [ ] `internal/adapters/primary/http/router.go`
  - Registra todas as rotas com Chi
  - Aplica middlewares globais (Logger, Recoverer, CORS)
- [ ] `internal/adapters/primary/http/handlers/showtime_handler.go`
  - `GET /showtimes?date=&movie_id=`
  - `GET /showtimes/:id`
  - `GET /showtimes/:id/seats`
  - `POST /showtimes`, `PUT /showtimes/:id`, `DELETE /showtimes/:id` (admin)
- [ ] `internal/adapters/primary/http/handlers/theater_handler.go`
  - `GET /theaters` (admin)
  - `POST /theaters` (admin)
  - `GET /theaters/:id/seats`

**Injeção de dependências:**

- [ ] Atualizar `cmd/api/main.go` com wire-up completo:
  ```
  main → config → db → repositories → services → handlers → router → server
  ```

### Critério de conclusão

Após merge das branches de ambas, `make run` sobe a API completa. Todos os endpoints respondem (testado via curl ou Postman).

---

## Etapa 4 — Reservas e Testes

> **Estimativa:** 2–3 dias  
> **Pré-requisito:** Etapa 3

### Dev 1 — Branch: `feature/tests-auth-movies`

Escrever testes unitários com mocks das interfaces de repositório:

- [ ] `internal/core/services/auth_service_test.go`
  - SignUp com email duplicado retorna erro
  - Login com senha errada retorna erro
  - Login correto retorna token JWT válido
- [ ] `internal/core/services/movie_service_test.go`
  - Criar filme sem gêneros válidos retorna erro
  - Deletar filme com sessões futuras retorna erro
- [ ] `internal/core/services/user_service_test.go`
  - Promover usuário que não existe retorna ErrNotFound

### Dev 2 — Branch: `feature/reservation-service`

Esta é a parte mais crítica do projeto — **prevenção de overbooking**.

- [ ] `internal/core/services/reservation_service.go`
  ```
  CreateReservation(ctx, userID, showtimeID, seatIDs):
    1. Iniciar transação no banco
    2. SELECT FOR UPDATE nos assentos solicitados
    3. Verificar se já existem em reservation_seats para o showtime
    4. Se livre: criar reservation + reservation_seats
    5. Calcular total_price = len(seats) * showtime.price
    6. Commit da transação
    7. Se erro de constraint UNIQUE: retornar ErrSeatAlreadyTaken
  ```

- [ ] `internal/core/services/reservation_service.go` — `CancelReservation`
  ```
  CancelReservation(ctx, userID, reservationID):
    1. Buscar reservation
    2. Verificar se pertence ao userID (ou se é admin)
    3. Verificar se showtime.start_time > time.Now()
    4. Atualizar status para "cancelled"
    5. Remover registros de reservation_seats
  ```

- [ ] `internal/adapters/primary/http/handlers/reservation_handler.go`
  - `GET /reservations/me`
  - `POST /reservations`
  - `DELETE /reservations/:id`
  - `GET /admin/reservations` (admin)

- [ ] `internal/core/services/reservation_service_test.go`
  - Reserva de assento já ocupado retorna ErrSeatAlreadyTaken
  - Cancelamento de reserva passada retorna ErrCannotCancelPastReservation
  - Cancelamento por usuário diferente retorna ErrForbidden

### Critério de conclusão

Teste de concorrência: 10 goroutines tentando reservar o mesmo assento simultaneamente — apenas 1 deve ter sucesso.

---

## Etapa 5 — Relatórios, Seed e Documentação

> **Estimativa:** 1–2 dias  
> **Pré-requisito:** Etapa 4

### Dev 1 — Branch: `feature/reports-revenue`

- [ ] Query de receita em `reservation_repository.go`
  ```sql
  SELECT m.title, s.start_time, COUNT(rs.seat_id) as tickets_sold,
         SUM(s.price) as revenue
  FROM showtimes s
  JOIN movies m ON m.id = s.movie_id
  JOIN reservations r ON r.showtime_id = s.id AND r.status = 'active'
  JOIN reservation_seats rs ON rs.reservation_id = r.id
  GROUP BY m.title, s.id
  ORDER BY s.start_time DESC
  ```

- [ ] `GET /admin/reports/revenue` com filtros opcionais: `?from=&to=&movie_id=`
- [ ] `scripts/seed.go` — dados para demonstração:
  - 1 admin (credenciais do `.env`)
  - 5 gêneros
  - 10 filmes com gêneros
  - 3 salas com assentos
  - Sessões para a próxima semana

### Dev 2 — Branch: `feature/reports-capacity`

- [ ] Query de capacidade em `showtime_repository.go`
  ```sql
  SELECT s.id, m.title, s.start_time,
         t.total_rows * t.seats_per_row as total_seats,
         COUNT(rs.seat_id) as reserved_seats,
         ROUND(COUNT(rs.seat_id)::numeric /
               (t.total_rows * t.seats_per_row) * 100, 2) as occupancy_pct
  FROM showtimes s
  JOIN movies m ON m.id = s.movie_id
  JOIN theaters t ON t.id = s.theater_id
  LEFT JOIN reservations r ON r.showtime_id = s.id AND r.status = 'active'
  LEFT JOIN reservation_seats rs ON rs.reservation_id = r.id
  GROUP BY s.id, m.title, s.start_time, t.total_rows, t.seats_per_row
  ```

- [ ] `GET /admin/reports/capacity`
- [ ] **Postman Collection** (`docs/movie-reservation.postman_collection.json`)
  - Todos os endpoints com exemplos de request/response
  - Variáveis de ambiente: `{{base_url}}`, `{{token}}`, `{{admin_token}}`

### Critério de conclusão

`make seed` popula o banco. Todos os endpoints de relatório retornam dados corretos.

---

## Etapa 6 — Integração Final e Deploy

> **Responsáveis:** Ambas  
> **Estimativa:** 1–2 dias  
> **Branch:** `feature/final-integration` → PR para `main`

### Checklist conjunto

- [ ] Revisão cruzada: Dev 1 revisa código da Dev 2 e vice-versa
- [ ] Testes de integração com banco real (Docker)
  - Fluxo completo: signup → login → browse movies → reserve → cancel
  - Fluxo admin: login admin → create movie → create showtime → view reports
- [ ] `Dockerfile` multi-stage otimizado
  ```dockerfile
  FROM golang:1.22-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN go build -o bin/api ./cmd/api

  FROM alpine:3.19
  WORKDIR /app
  COPY --from=builder /app/bin/api .
  COPY --from=builder /app/migrations ./migrations
  EXPOSE 8080
  CMD ["./api"]
  ```
- [ ] `docker-compose.yml` com healthcheck no postgres e depends_on na api
- [ ] Variáveis de produção documentadas no `.env.example`
- [ ] Atualizar README com instruções finais
- [ ] Tag `v1.0.0` no git

---

## Fluxo de Trabalho com Git

### Criando uma feature

```bash
# Sempre a partir de develop atualizado
git checkout develop
git pull origin develop
git checkout -b feature/minha-feature

# Trabalhe, commite com frequência
git add .
git commit -m "feat: descrição do que foi feito"

# Antes de abrir PR, atualize com develop
git fetch origin
git rebase origin/develop

# Abra PR: feature/minha-feature → develop
```

### Resolvendo conflitos

```bash
# Durante o rebase, se houver conflito:
# 1. Resolva os arquivos com conflito
# 2. git add arquivo-resolvido
# 3. git rebase --continue
```

### Merge para main (somente releases)

```bash
# PR: develop → main (requer aprovação das duas)
# Após merge, criar tag:
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

---

## Dependências entre Etapas

```
Etapa 0 (ambas)
    └─► Etapa 1 Dev1 ──► Etapa 2 Dev1 ──► Etapa 3 Dev1 ──► Etapa 4 Dev1 ──► Etapa 5 Dev1
    └─► Etapa 1 Dev2 ──► Etapa 2 Dev2 ──► Etapa 3 Dev2 ──► Etapa 4 Dev2 ──► Etapa 5 Dev2
                                                                                    └─► Etapa 6 (ambas)
```

**Pontos de sincronização obrigatória:**
1. Após Etapa 0 — revisar estrutura juntas antes de começar paralelo
2. Após Etapa 3 — merge das duas branches antes da Etapa 4 (reservation_handler precisa dos middlewares de Dev 1)
3. Antes da Etapa 6 — todas as features em `develop`, build passando

---

## Estimativa Total

| Etapa | Dias |
|---|---|
| 0 — Setup | 1 |
| 1 — Domain + Ports | 1–2 |
| 2 — Services + Repos | 2–3 |
| 3 — HTTP Handlers | 2 |
| 4 — Reservas + Testes | 2–3 |
| 5 — Relatórios + Docs | 1–2 |
| 6 — Integração Final | 1–2 |
| **Total** | **~10–15 dias úteis** |
