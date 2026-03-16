# Guia de Desenvolvimento — Movie Reservation System

Este documento orienta o próximo dev a iniciar a **Etapa 1** a partir da base já entregue no **Setup 0**.

---

## Status Atual

> Última atualização: Etapa 0 concluída — **Etapa 1 é a próxima para ambas**

| Etapa | Status | Responsável |
|---|---|---|
| **0 — Setup** | ✅ Concluída | Ambas |
| **1 — Domain + Ports** | ⬜ Pendente — próxima | Dev 1 + Dev 2 em paralelo |
| **2 — Services + Repos** | ⬜ Pendente | Dev 1 + Dev 2 em paralelo |
| **3 — HTTP Handlers** | ⬜ Pendente | Dev 1 + Dev 2 em paralelo |
| **4 — Reservas + Testes** | ⬜ Pendente | Dev 1 + Dev 2 em paralelo |
| **5 — Relatórios + Docs** | ⬜ Pendente | Dev 1 + Dev 2 em paralelo |
| **6 — Integração Final** | ⬜ Pendente | Ambas |

---

## O que já existe no repositório (Etapa 0)

Antes de começar a Etapa 1, toda a infraestrutura base já está criada:

```
go.mod                                      ← módulo + 9 dependências declaradas
.env.example                                ← variáveis de ambiente documentadas
.gitignore
Dockerfile                                  ← multi-stage build
docker-compose.yml                          ← PostgreSQL 15 + API + healthcheck
Makefile                                    ← 15 comandos prontos
cmd/api/main.go                             ← servidor HTTP + graceful shutdown
internal/adapters/primary/http/router.go    ← Chi + middlewares globais + GET /health
pkg/config/config.go                        ← leitura de .env
pkg/logger/logger.go                        ← slog estruturado
pkg/apperrors/errors.go                     ← erros de domínio tipados
pkg/auth/jwt.go                             ← GenerateToken + ValidateToken
migrations/000001_create_users.*
migrations/000002_create_movies.*           ← inclui genres e movie_genres
migrations/000003_create_theaters.*         ← inclui seats
migrations/000004_create_showtimes.*
migrations/000005_create_reservations.*     ← inclui UNIQUE(seat_id, showtime_id)
scripts/seed.go                             ← admin + gêneros + filmes + salas
```

### Antes de começar a Etapa 1

```bash
# 1. Instalar Go 1.22+  →  https://go.dev/dl/

# 2. Clonar o repositório
git clone https://github.com/<usuario>/movie-reservation.git
cd movie-reservation

# 3. Copiar o .env e definir o JWT_SECRET
cp .env.example .env
# Editar .env: trocar JWT_SECRET por um valor forte
# Exemplo: openssl rand -base64 32

# 4. Baixar dependências e gerar go.sum
make tidy

# 5. Confirmar que o projeto compila
go build ./...

# 6. Subir o banco, rodar migrations e seed
make docker-db
make migrate-up
make seed

# 7. Rodar a API e testar o health check
make run
curl http://localhost:8080/health
# Esperado: {"env":"development","status":"ok"}
```

---

## Divisão de Domínios

| | Dev 1 | Dev 2 |
|---|---|---|
| **Domínio principal** | Auth · Usuários · Filmes · Gêneros | Salas · Assentos · Sessões · Reservas · Relatórios |
| **Camadas** | domain → ports → service → adapter | domain → ports → service → adapter |

> **Regra de ouro:** o `core` nunca importa nada dos `adapters`. Dependência sempre de fora para dentro.

---

## Etapa 0 — Setup do Projeto ✅

> **Responsáveis:** Ambas
> **Status: CONCLUÍDA**

### O que foi feito

**Infraestrutura:**
- `go.mod` com módulo `github.com/movie-reservation/api` e 9 dependências
- `Dockerfile` multi-stage: builder `golang:1.22-alpine` → runtime `alpine:3.19`
- `docker-compose.yml`: PostgreSQL 15 com healthcheck; serviço `api` com `depends_on: postgres`
- `.env.example` com todas as variáveis necessárias
- `.gitignore` cobrindo binários, `.env`, cobertura e editores
- `Makefile` com 15 comandos prontos para uso

**Pacotes base (`pkg/`):**
- `pkg/config/config.go` — lê `.env` e variáveis de ambiente; `DSN()` monta a string de conexão; `requireEnv` faz panic se variável obrigatória estiver ausente
- `pkg/logger/logger.go` — `slog`: JSON em `production`, texto em `development`
- `pkg/apperrors/errors.go` — erros tipados com `StatusCode`, `Code` e `Unwrap`; inclui `ErrSeatAlreadyTaken`, `ErrCannotCancelPastReservation` e outros
- `pkg/auth/jwt.go` — `GenerateToken` e `ValidateToken` com `golang-jwt/jwt v5`

**Servidor HTTP:**
- `internal/adapters/primary/http/router.go` — Chi com middlewares globais (`RequestID`, `RealIP`, `Logger`, `Recoverer`, `Timeout`) e `GET /health`
- `cmd/api/main.go` — wire-up `config → logger → router → http.Server`; graceful shutdown em `SIGINT`/`SIGTERM` com 30s de timeout

**Banco de dados (migrations):**

| Arquivo | Tabelas criadas |
|---|---|
| `000001_create_users` | `users` + enum `user_role` + extensão `uuid-ossp` |
| `000002_create_movies` | `movies`, `genres`, `movie_genres` (N:N) |
| `000003_create_theaters` | `theaters`, `seats` — `UNIQUE(theater_id, row, number)` |
| `000004_create_showtimes` | `showtimes` — constraint `end_time > start_time` |
| `000005_create_reservations` | `reservations`, `reservation_seats` — `UNIQUE(seat_id, showtime_id)` |

**Seed (`scripts/seed.go`):**
- 1 usuário admin (credenciais do `.env`)
- 5 gêneros
- 3 filmes com gêneros associados
- 3 salas com assentos gerados automaticamente (A1…J15 por exemplo)

### Notas técnicas

- O pacote `internal/adapters/primary/http/` é declarado como `package server` para evitar conflito com `net/http`
- `scripts/seed.go` usa `//go:build ignore` — nunca entra no binário; execute com `go run ./scripts/seed.go`
- O `go.sum` ainda não existe — será gerado automaticamente com `make tidy` após instalar o Go

---

## Etapa 1 — Entidades de Domínio e Ports

> **Estimativa:** 1–2 dias
> **Pré-requisito:** Etapa 0 + `make tidy` executado com sucesso

### Dev 1 — Branch: `feature/domain-auth-movies`

**Entidades de domínio:**

- [ ] `internal/core/domain/user.go`
  ```go
  type Role string
  const (
      RoleAdmin Role = "admin"
      RoleUser  Role = "user"
  )

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

**Ports de entrada (use cases):**

- [ ] `internal/core/ports/inbound/auth_service.go`
  ```go
  type AuthService interface {
      SignUp(ctx context.Context, input SignUpInput) (*domain.User, error)
      Login(ctx context.Context, input LoginInput) (string, error)
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

**Ports de saída (repositórios):**

- [ ] `internal/core/ports/outbound/user_repository.go`
  ```go
  type UserRepository interface {
      Create(ctx context.Context, user *domain.User) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
      GetByEmail(ctx context.Context, email string) (*domain.User, error)
      List(ctx context.Context) ([]domain.User, error)
      Update(ctx context.Context, user *domain.User) error
  }
  ```

- [ ] `internal/core/ports/outbound/movie_repository.go`
  ```go
  type MovieRepository interface {
      Create(ctx context.Context, movie *domain.Movie) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Movie, error)
      List(ctx context.Context, filters MovieFilters) ([]domain.Movie, error)
      Update(ctx context.Context, movie *domain.Movie) error
      Delete(ctx context.Context, id uuid.UUID) error
  }
  ```

### Dev 2 — Branch: `feature/domain-showtime-reservation`

**Entidades de domínio:**

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
  const (
      StatusActive    ReservationStatus = "active"
      StatusCancelled ReservationStatus = "cancelled"
  )

  type Reservation struct {
      ID         uuid.UUID
      UserID     uuid.UUID
      ShowtimeID uuid.UUID
      Seats      []Seat
      TotalPrice float64
      Status     ReservationStatus
      CreatedAt  time.Time
      UpdatedAt  time.Time
  }
  ```

**Ports de entrada (use cases):**

- [ ] `internal/core/ports/inbound/showtime_service.go`
  ```go
  type ShowtimeService interface {
      Create(ctx context.Context, input CreateShowtimeInput) (*domain.Showtime, error)
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Showtime, error)
      List(ctx context.Context, filters ShowtimeFilters) ([]domain.Showtime, error)
      GetAvailableSeats(ctx context.Context, showtimeID uuid.UUID) ([]domain.Seat, error)
      Update(ctx context.Context, id uuid.UUID, input UpdateShowtimeInput) (*domain.Showtime, error)
      Delete(ctx context.Context, id uuid.UUID) error
  }
  ```

- [ ] `internal/core/ports/inbound/reservation_service.go`
  ```go
  type ReservationService interface {
      Create(ctx context.Context, input CreateReservationInput) (*domain.Reservation, error)
      Cancel(ctx context.Context, userID, reservationID uuid.UUID) error
      ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Reservation, error)
      ListAll(ctx context.Context) ([]domain.Reservation, error)
  }
  ```

**Ports de saída (repositórios):**

- [ ] `internal/core/ports/outbound/theater_repository.go`
  ```go
  type TheaterRepository interface {
      Create(ctx context.Context, theater *domain.Theater) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Theater, error)
      List(ctx context.Context) ([]domain.Theater, error)
      GetSeats(ctx context.Context, theaterID uuid.UUID) ([]domain.Seat, error)
  }
  ```

- [ ] `internal/core/ports/outbound/showtime_repository.go`
  ```go
  type ShowtimeRepository interface {
      Create(ctx context.Context, showtime *domain.Showtime) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Showtime, error)
      List(ctx context.Context, filters ShowtimeFilters) ([]domain.Showtime, error)
      GetAvailableSeats(ctx context.Context, showtimeID uuid.UUID) ([]domain.Seat, error)
      Update(ctx context.Context, showtime *domain.Showtime) error
      Delete(ctx context.Context, id uuid.UUID) error
  }
  ```

- [ ] `internal/core/ports/outbound/reservation_repository.go`
  ```go
  type ReservationRepository interface {
      Create(ctx context.Context, reservation *domain.Reservation) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Reservation, error)
      ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Reservation, error)
      ListAll(ctx context.Context) ([]domain.Reservation, error)
      Cancel(ctx context.Context, id uuid.UUID) error
      IsSeatsAvailable(ctx context.Context, showtimeID uuid.UUID, seatIDs []uuid.UUID) (bool, error)
  }
  ```

### Critério de conclusão

`go build ./...` passa sem erros. Nenhum arquivo de domínio importa pacotes de infraestrutura.

---

## Fluxo de Trabalho com Git

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
