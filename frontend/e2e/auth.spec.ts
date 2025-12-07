import { test, expect } from '@playwright/test';

test('login page loads successfully', async ({ page }) => {
    // Use the QA URL or localhost if running locally
    // In CI, this will be set to the QA URL
    const baseURL = process.env.PLAYWRIGHT_TEST_BASE_URL || 'http://localhost:3000';

    await page.goto(baseURL);

    // Check for the "Inicia Sesión" text which indicates the login page is loaded
    await expect(page.getByText('Inicia Sesión')).toBeVisible();

    // Check for email and password inputs
    await expect(page.getByPlaceholder('tu@email.com')).toBeVisible();
    await expect(page.getByPlaceholder('••••••••')).toBeVisible();
});
