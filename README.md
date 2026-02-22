# KahootClone — Real-Time Quiz App

A production-grade, real-time quiz application similar to Kahoot, built with Go (backend) and React + TypeScript (frontend). Designed for AWS serverless deployment (Lambda, DynamoDB, ElastiCache, Cognito) with a unified local development server.

## Architecture

```
┌─────────────┐       WebSocket        ┌──────────────────┐
│   React UI  │ ◄────────────────────► │  Go Backend      │
│   (Vite)    │       REST API         │  (local server   │
│             │ ◄────────────────────► │   or Lambda)     │
└─────────────┘                        └────────┬─────────┘
                                                │
                              ┌─────────────────┼─────────────────┐
                              │                 │                 │
                         ┌────▼────┐      ┌─────▼─────┐    ┌─────▼─────┐
                         │ DynamoDB│      │   Redis   │    │  Cognito  │
                         │         │      │(Leaderboard)│   │  (Auth)   │
                         └─────────┘      └───────────┘    └───────────┘
```

## Tech Stack

### Backend

- **Language**: Go 1.22+
- **Database**: Amazon DynamoDB
- **Cache**: Amazon ElastiCache for Redis
- **Auth**: Amazon Cognito JWT
- **WebSocket**: gorilla/websocket (local) / API Gateway WebSocket (prod)
- **Observability**: AWS X-Ray + structured logging (slog)

### Frontend

- **Framework**: React 18 + TypeScript + Vite
- **State**: Zustand
- **Auth**: amazon-cognito-identity-js
- **UI**: Tailwind CSS v3 + shadcn/ui
- **Animations**: Framer Motion

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+
- DynamoDB Local (optional)
- Redis

### Backend

```bash
cd backend
cp .env.example .env   # Edit with your values
go run ./cmd/local
```

### Frontend

```bash
cd frontend
cp .env.example .env   # Edit with your values
npm install
npm run dev
```

### Using Make

```bash
make dev-backend    # Start Go local server
make dev-frontend   # Start Vite dev server
make build-lambda   # Compile all Lambda handlers
make tidy           # go mod tidy + npm install
```

## Project Structure

```
kahootclone/
├── backend/
│   ├── cmd/
│   │   ├── local/           # Local dev server
│   │   └── lambda/          # Lambda handlers (9 functions)
│   ├── internal/
│   │   ├── auth/            # Cognito JWT validation
│   │   ├── db/              # DynamoDB operations
│   │   ├── cache/           # Redis leaderboard
│   │   ├── models/          # Data types
│   │   ├── game/            # Engine, scoring, broadcast
│   │   ├── observability/   # Logger + tracer
│   │   └── config/          # Env var config
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/             # Axios client
│   │   ├── auth/            # Cognito auth
│   │   ├── websocket/       # WS hook
│   │   ├── store/           # Zustand stores
│   │   ├── pages/           # 11 pages
│   │   └── components/      # UI components
│   └── package.json
├── Makefile
└── README.md
```
