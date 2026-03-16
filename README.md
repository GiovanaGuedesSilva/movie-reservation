# Movie Reservation System

Sistema de reserva de ingressos de cinema com autenticação, gerenciamento de filmes, sessões e assentos.

Projeto baseado no desafio [roadmap.sh — Movie Reservation System](https://roadmap.sh/projects/movie-reservation-system).

---

## Sumário

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Tech Stack](#tech-stack)
- [Estrutura de Pastas](#estrutura-de-pastas)
- [Modelo de Dados](#modelo-de-dados)
- [Endpoints da API](#endpoints-da-api)
- [Pré-requisitos](#pré-requisitos)
- [Configuração do Ambiente](#configuração-do-ambiente)
- [Executando o Projeto](#executando-o-projeto)
- [Comandos Makefile](#comandos-makefile)
- [Testes](#testes)
- [Convenções](#convenções)

---

## Visão Geral

O sistema permite que:

- **Usuários** se cadastrem, façam login, consultem filmes e sessões, reservem assentos e gerenciem suas reservas.
- **Administradores** gerenciem filmes, gêneros, salas, sessões e visualizem relatórios de receita e capacidade.

### Funcionalidades

| Módulo | Funcionalidade |
|---|---|
| Auth | Cadastro, login, JWT, roles (admin/user) |
| Usuários | Perfil, promoção para admin |
| Filmes | CRUD completo, categorização por gênero |
| Salas | Gerenciamento de salas e assentos |
| Sessões | Agendamento de sessões por sala/filme |
| Reservas | Reserva de assentos, cancelamento, histórico |
| Relatórios | Receita, capacidade, listagem admin |

---

## Arquitetura

O projeto segue a **Arquitetura Hexagonal** (Ports and Adapters), que isola o núcleo de negócio de detalhes de infraestrutura (banco de dados, HTTP, etc.).

```
┌─────────────────────────────────────────────────────────┐
│                      ADAPTERS                           │
│                                                         │
│   ┌─────────────────┐         ┌─────────────────────┐   │
│   │  Primary (HTTP) │         │ Secondary (Postgres)│   │
│   │  Chi Handlers   │         │  GORM Repositories  │   │
│   └────────┬────────┘         └──────────┬──────────┘   │
│            │ inbound ports               │outbound ports│
├────────────▼─────────────────────────────▼──────────────┤
│                        CORE                             │
│                                                         │
│   ┌──────────────────────────────────────────────────┐  │
│   │                    SERVICES                      │  │
│   │  AuthService | MovieService | ReservationService │  │
│   └──────────────────────┬───────────────────────────┘  │
│                          │                              │
│   ┌──────────────────────▼───────────────────────────┐  │
│   │                    DOMAIN                        │  │
│   │  User | Movie | Genre | Showtime | Reservation   │  │
│   └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**Regra principal:** o `core` não importa nada dos `adapters`. A dependência é sempre de fora para dentro.

---

## Tech Stack

| Componente | Tecnologia | Motivo |
|---|---|---|
| Linguagem | Go 1.22+ | Performance, tipagem forte, concorrência nativa |
| HTTP Router | [Chi v5](https://github.com/go-chi/chi) | Leve, idiomático, middleware-friendly |
| ORM | [GORM](https://gorm.io) | Produtividade, migrations automáticas |
| Banco de Dados | PostgreSQL 15 | Relacional, suporte a transações, ideal para reservas |
| Autenticação | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) | Padrão de mercado para JWT |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) | Controle de versão do schema |
| Configuração | [godotenv](https://github.com/joho/godotenv) | Leitura de `.env` |
| Testes | [testify](https://github.com/stretchr/testify) | Assertions claras, mocks |
| Containers | Docker + Docker Compose | Ambiente reproduzível |
| Criptografia | bcrypt (`golang.org/x/crypto`) | Hash seguro de senhas |

---

## Estrutura de Pastas

```
movie-reservation/
│
├── cmd/
│   └── api/
│       └── main.go                    # Ponto de entrada da aplicação
│
├── internal/
│   ├── core/
│   │   ├── domain/                    # Entidades de domínio (sem dependências externas)
│   │   │   ├── user.go
│   │   │   ├── movie.go
│   │   │   ├── genre.go
│   │   │   ├── theater.go
│   │   │   ├── seat.go
│   │   │   ├── showtime.go
│   │   │   └── reservation.go
│   │   │
│   │   ├── ports/
│   │   │   ├── inbound/               # Contratos de entrada (use cases)
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── user_service.go
│   │   │   │   ├── movie_service.go
│   │   │   │   ├── showtime_service.go
│   │   │   │   └── reservation_service.go
│   │   │   │
│   │   │   └── outbound/              # Contratos de saída (repositórios)
│   │   │       ├── user_repository.go
│   │   │       ├── movie_repository.go
│   │   │       ├── theater_repository.go
│   │   │       ├── showtime_repository.go
│   │   │       └── reservation_repository.go
│   │   │
│   │   └── services/                  # Implementação da lógica de negócio
│   │       ├── auth_service.go
│   │       ├── user_service.go
│   │       ├── movie_service.go
│   │       ├── showtime_service.go
│   │       └── reservation_service.go
│   │
│   └── adapters/
│       ├── primary/
│       │   └── http/
│       │       ├── router.go          # Registro de rotas
│       │       ├── middleware/
│       │       │   ├── auth.go        # Extrai e valida JWT
│       │       │   └── role.go        # Verifica permissão admin
│       │       └── handlers/
│       │           ├── auth_handler.go
│       │           ├── user_handler.go
│       │           ├── movie_handler.go
│       │           ├── showtime_handler.go
│       │           └── reservation_handler.go
│       │
│       └── secondary/
│           └── postgres/
│               ├── db.go              # Conexão GORM
│               ├── models.go          # Modelos GORM (separados do domain)
│               ├── user_repository.go
│               ├── movie_repository.go
│               ├── theater_repository.go
│               ├── showtime_repository.go
│               └── reservation_repository.go
│
├── pkg/
│   ├── auth/
│   │   └── jwt.go                     # Geração e validação de tokens
│   ├── config/
│   │   └── config.go                  # Leitura de variáveis de ambiente
│   ├── apperrors/
│   │   └── errors.go                  # Erros de domínio tipados (ErrNotFound, etc.)
│   └── logger/
│       └── logger.go                  # Logger estruturado
│
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_movies.up.sql
│   └── ...
│
├── scripts/
│   └── seed.go                        # Cria admin inicial e dados de teste
│
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── DEV.md
```

---

## Modelo de Dados

### Diagrama de Relacionamentos

```
users
  id, email, password_hash, name, role (admin|user), created_at, updated_at

movies
  id, title, description, poster_url, duration_minutes, created_at, updated_at

genres
  id, name

movie_genres (N:N)
  movie_id → movies.id
  genre_id → genres.id

theaters
  id, name, total_rows, seats_per_row, created_at

seats
  id, theater_id → theaters.id
  row (A, B, C...), number (1, 2, 3...)
  UNIQUE(theater_id, row, number)

showtimes
  id, movie_id → movies.id, theater_id → theaters.id
  start_time, end_time, price, created_at

reservations
  id, user_id → users.id, showtime_id → showtimes.id
  status (active|cancelled), total_price, created_at, updated_at

reservation_seats
  reservation_id → reservations.id
  seat_id        → seats.id
  showtime_id    → showtimes.id
  UNIQUE(seat_id, showtime_id)   ← previne overbooking
```

### Prevenção de Overbooking

A constraint `UNIQUE(seat_id, showtime_id)` na tabela `reservation_seats` garante, em nível de banco de dados, que um assento não pode ser reservado duas vezes na mesma sessão. O service de reserva também usa **transação com SELECT FOR UPDATE** para evitar race conditions em requisições concorrentes.

---

## Endpoints da API

Base URL: `http://localhost:8080/api/v1`

### Autenticação

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `POST` | `/auth/signup` | Público | Cadastro de usuário |
| `POST` | `/auth/login` | Público | Login, retorna JWT |

**POST /auth/signup**
```json
{
  "name": "Maria Silva",
  "email": "maria@email.com",
  "password": "senha123"
}
```

**POST /auth/login**
```json
{
  "email": "maria@email.com",
  "password": "senha123"
}
```
Resposta:
```json
{
  "token": "eyJhbGci...",
  "user": { "id": "uuid", "name": "Maria Silva", "role": "user" }
}
```

---

### Usuários

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `GET` | `/users/me` | Autenticado | Dados do usuário logado |
| `PUT` | `/users/:id/promote` | Admin | Promove usuário a admin |
| `GET` | `/users` | Admin | Lista todos os usuários |

---

### Filmes e Gêneros

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `GET` | `/movies` | Público | Lista filmes (com filtros) |
| `GET` | `/movies/:id` | Público | Detalhe do filme |
| `POST` | `/movies` | Admin | Cria filme |
| `PUT` | `/movies/:id` | Admin | Atualiza filme |
| `DELETE` | `/movies/:id` | Admin | Remove filme |
| `GET` | `/genres` | Público | Lista gêneros |
| `POST` | `/genres` | Admin | Cria gênero |

**POST /movies**
```json
{
  "title": "Duna: Parte 2",
  "description": "A continuação da saga...",
  "poster_url": "https://...",
  "duration_minutes": 166,
  "genre_ids": ["uuid-acao", "uuid-ficcao"]
}
```

---

### Salas e Assentos

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `GET` | `/theaters` | Admin | Lista salas |
| `POST` | `/theaters` | Admin | Cria sala (gera assentos automaticamente) |
| `GET` | `/theaters/:id/seats` | Autenticado | Lista assentos da sala |

**POST /theaters**
```json
{
  "name": "Sala 1 - IMAX",
  "total_rows": 10,
  "seats_per_row": 15
}
```
> Os 150 assentos (A1–J15) são gerados automaticamente.

---

### Sessões

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `GET` | `/showtimes` | Público | Lista sessões (`?date=2026-03-16&movie_id=...`) |
| `GET` | `/showtimes/:id` | Público | Detalhe da sessão |
| `GET` | `/showtimes/:id/seats` | Autenticado | Assentos disponíveis/ocupados |
| `POST` | `/showtimes` | Admin | Cria sessão |
| `PUT` | `/showtimes/:id` | Admin | Atualiza sessão |
| `DELETE` | `/showtimes/:id` | Admin | Remove sessão |

**POST /showtimes**
```json
{
  "movie_id": "uuid",
  "theater_id": "uuid",
  "start_time": "2026-03-20T19:00:00Z",
  "price": 35.00
}
```

---

### Reservas

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `GET` | `/reservations/me` | Autenticado | Reservas do usuário logado |
| `POST` | `/reservations` | Autenticado | Cria reserva |
| `DELETE` | `/reservations/:id` | Autenticado | Cancela reserva (somente futuras) |
| `GET` | `/admin/reservations` | Admin | Todas as reservas |

**POST /reservations**
```json
{
  "showtime_id": "uuid",
  "seat_ids": ["uuid-A1", "uuid-A2"]
}
```
Resposta:
```json
{
  "id": "uuid",
  "showtime_id": "uuid",
  "seats": [{ "row": "A", "number": 1 }, { "row": "A", "number": 2 }],
  "total_price": 70.00,
  "status": "active"
}
```

---

### Relatórios (Admin)

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| `GET` | `/admin/reports/revenue` | Admin | Receita total por filme/sessão |
| `GET` | `/admin/reports/capacity` | Admin | Taxa de ocupação por sessão |

---

## Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) e Docker Compose
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (opcional, tem no Makefile)
- [Make](https://www.gnu.org/software/make/) (Linux/macOS nativo; Windows: [GnuWin32](http://gnuwin32.sourceforge.net/packages/make.htm) ou WSL)

---

## Configuração do Ambiente

```bash
# 1. Clone o repositório
git clone https://github.com/seu-usuario/movie-reservation.git
cd movie-reservation

# 2. Copie o arquivo de variáveis de ambiente
cp .env.example .env

# 3. Edite o .env com suas configurações locais
```

**.env.example**
```env
# Servidor
APP_PORT=8080
APP_ENV=development

# Banco de Dados
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=movie_reservation
DB_SSLMODE=disable

# JWT
JWT_SECRET=troque_por_um_segredo_forte
JWT_EXPIRY_HOURS=24

# Seed
ADMIN_EMAIL=admin@cinema.com
ADMIN_PASSWORD=Admin@123
```

---

## Executando o Projeto

### Com Docker (recomendado)

```bash
# Sobe o banco e a aplicação
make docker-up

# Para tudo
make docker-down
```

### Localmente (Go instalado)

```bash
# 1. Sobe apenas o banco
docker-compose up -d postgres

# 2. Instala dependências
make install

# 3. Roda as migrations
make migrate-up

# 4. Popula o banco com dados iniciais (admin + dados de teste)
make seed

# 5. Inicia o servidor
make run
```

O servidor estará disponível em `http://localhost:8080`.

---

## Comandos Makefile

| Comando | Descrição |
|---|---|
| `make run` | Inicia o servidor Go |
| `make build` | Compila o binário em `bin/api` |
| `make test` | Roda todos os testes |
| `make test-coverage` | Testes com relatório de cobertura |
| `make install` | Baixa dependências (`go mod download`) |
| `make migrate-up` | Aplica todas as migrations |
| `make migrate-down` | Reverte a última migration |
| `make migrate-create NAME=foo` | Cria nova migration |
| `make seed` | Popula o banco com dados iniciais |
| `make docker-up` | Sobe todos os containers |
| `make docker-down` | Para todos os containers |
| `make lint` | Roda o linter (golangci-lint) |

---

## Testes

```bash
# Todos os testes
make test

# Testes de um pacote específico
go test ./internal/core/services/...

# Com cobertura
make test-coverage
```

Os testes de serviço usam **mocks das interfaces de repositório** (sem banco real). Os testes de integração sobem um banco PostgreSQL via Docker.

---

## Convenções

### Branches

```
main          → produção, protegida
develop       → integração contínua
feature/nome  → nova funcionalidade (ex: feature/auth-jwt)
fix/nome      → correção de bug
```

### Commits (Conventional Commits)

```
feat: adiciona endpoint de reserva de assentos
fix: corrige validação de sessões passadas no cancelamento
refactor: extrai lógica de JWT para pkg/auth
test: adiciona testes do reservation_service
docs: atualiza endpoints no README
```

### Código Go

- **Interfaces** definidas no pacote que **consome**, não no que implementa.
- **Erros de domínio** tipados em `pkg/apperrors` (ex: `apperrors.ErrNotFound`).
- **Handlers** só traduzem HTTP → domain e domain → HTTP. Zero lógica de negócio.
- **Services** nunca importam `net/http` ou qualquer pacote de infraestrutura.
- Nomes em **inglês** no código; comentários e PR descriptions podem ser em português.

### Respostas da API

Sucesso:
```json
{ "data": { ... } }
```

Erro:
```json
{ "error": "mensagem legível", "code": "ERR_CODE" }
```
