# Gestor-Gastos

## Backend (Go + Gin)

1. Instala dependencias (solo la primera vez): `cd backend && go mod tidy`.
2. Inicia la API: `go run .`.
3. La API expone:
   - `POST /api/auth/register`, `POST /api/auth/login`, `GET /api/auth/me`.
   - CRUD básico para `/api/expenses` y `/api/monthly-expenses`.

Cada arranque ejecuta las migraciones necesarias para crear las tablas `users`, `expenses` y `monthly_expenses` en la base QA de Supabase. El driver se conecta con `binary_parameters=yes` para deshabilitar prepared statements, requisito cuando usas el transaction pooler de Supabase (y así evitar errores como `bind message has ...`).
La carpeta `backend/` sigue una organización MVC clásica: `models/` concentra los esquemas y migraciones, `controllers/` contiene la lógica HTTP y `routes/` define los endpoints apoyados por los middlewares en `middleware/`.

## Frontend (Next.js)

1. `cd frontend && npm install`.
2. `npm run dev` y abre `http://localhost:3000`.
3. Autentícate (registro/login) para sincronizar los datos con el backend.
