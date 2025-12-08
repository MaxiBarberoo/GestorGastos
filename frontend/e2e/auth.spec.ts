import { test, expect } from '@playwright/test';

test('complete user journey: register, login, add expenses, verify total', async ({ page }) => {
    const baseURL = process.env.PLAYWRIGHT_TEST_BASE_URL || 'http://localhost:3000';
    const randomId = Math.floor(Math.random() * 10000);
    const userEmail = `test_user_${randomId}@example.com`;
    const userPassword = 'password123';
    const userName = `Test User ${randomId}`;

    console.log(`Starting test with user: ${userEmail}`);

    // 1. Navigate to home
    await page.goto(baseURL);

    // Wait for loading to finish
    await expect(page.getByText('Cargando...')).not.toBeVisible({ timeout: 10000 }).catch(() => { });

    // 2. Register new user
    await page.getByRole('button', { name: 'Regístrate' }).click();
    await expect(page.getByRole('heading', { name: 'Crea tu cuenta' })).toBeVisible();

    await page.getByLabel('Nombre').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByLabel('Contraseña').fill(userPassword);

    await page.getByRole('button', { name: 'Registrarme' }).click();

    // 3. Verify auto-login / Dashboard loaded
    // We expect to see "Resumen de Gastos" and the user's name
    await expect(page.getByText('Resumen de Gastos')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText(`Hola, ${userName}`)).toBeVisible();

    // 4. Add a regular expense
    // Name: Pizza, Tag: Comida, Amount: 1000
    await page.getByLabel('Nombre').first().fill('Pizza'); // .first() because there are two forms (regular and recurring)
    await page.getByPlaceholder('Escribe o selecciona').first().fill('Comida');
    await page.getByLabel('Monto').first().fill('1000');

    await page.getByRole('button', { name: 'Agregar', exact: true }).click();

    // Verify expense appears in list
    await expect(page.getByText('Pizza')).toBeVisible();
    await expect(page.getByText('$1000.00')).toBeVisible();

    // 5. Add a recurring expense
    // Name: Internet, Tag: Servicios, Amount: 5000
    // Note: recurring form is the second one
    await page.getByLabel('Nombre').nth(1).fill('Internet');
    await page.getByPlaceholder('Escribe o selecciona').nth(1).fill('Servicios');
    await page.getByLabel('Monto').nth(1).fill('5000');

    await page.getByRole('button', { name: 'Agregar Recurrente' }).click();

    // Verify recurring expense appears in list
    await expect(page.getByText('Internet')).toBeVisible();
    await expect(page.getByText('$5000.00')).toBeVisible();

    // 6. Apply recurring expense (to make it count in totals)
    await page.getByRole('button', { name: 'Aplicar' }).click();

    // 7. Verify Total
    // Expected: 1000 (Pizza) + 5000 (Internet) = 6000
    // The total is displayed as a large text, e.g., "$6000.00"
    await expect(page.getByText('$6000.00')).toBeVisible();

    console.log('Test completed successfully');
});
