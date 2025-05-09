# Backend Go com Google OAuth

Este projeto demonstra a implementação de um sistema de autenticação e autorização usando Golang com o framework Gin e integração com Google OAuth 2.0. O sistema inclui gerenciamento de permissões, roles e grupos de usuários, tudo dockerizado para fácil implantação.

## Características

- Autenticação com Google OAuth 2.0
- Geração e validação de JWT
- Sistema de permissões e roles
- Arquitetura em camadas (handlers, services, repositories)
- Suporte a PostgreSQL
- Dockerizado para fácil implantação

## Pré-requisitos

- Docker e Docker Compose
- Projeto configurado no Google Cloud Platform com OAuth
- (Opcional) Postman ou outra ferramenta para testar APIs

## Tecnologias Utilizadas
- **Backend**: Golang com Gin Framework
- **Autenticação**: Google OAuth 2.0
- **Tokens**: JWT (JSON Web Tokens)
- **Persistência**: PostgreSQL com GORM
- **Containerização**: Docker e Docker Compose
- **Arquitetura**: Camadas (handlers, services, repositories, models)


## Fluxo de Autenticação
```
┌──────────────┐       ┌────────────┐        ┌───────────────┐
│    Cliente   │       │  Backend   │        │  Google Auth  │
└──────┬───────┘       └──────┬─────┘        └───────┬───────┘
       │                      │                      │
       │   Inicia Login       │                      │
       │─────────────────────>│                      │
       │                      │                      │
       │                      │    Redireciona       │
       │<─────────────────────│─────────────────────>│
       │                      │                      │
       │      Autoriza        │                      │
       │─────────────────────>│                      │
       │                      │                      │
       │                      │   Token + Info User  │
       │                      │<─────────────────────│
       │                      │                      │
       │     JWT Tokens       │                      │
       │<─────────────────────│                      │
       │                      │                      │
```

## Principais Endpoints

### Autenticação
- `GET /auth/login` - Inicia fluxo de login
- `GET /auth/callback` - Callback do Google OAuth
- `POST /auth/refresh` - Renovação de tokens

### Usuários e Permissões
- `GET /api/profile` - Perfil do usuário autenticado
- `GET /api/admin/users` - Listar usuários (requer permissão admin)