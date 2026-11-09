
const { test, expect } = require('@playwright/test');

test('user can register, log in and log out', async ({ page }) => {
    // Unique user for each test run
    const username = `testuser_${Date.now()}`;
    const email = `${username}@example.com`;
    const password = 'TestPassword123!';

    // Register a new user
    await page.goto('/register');

    await page.locator('#username').fill(username);
    await page.locator('#email').fill(email);
    await page.locator('#password').fill(password);
    await page.locator('#password2').fill(password);

    await page.getByRole('button', { name: 'Sign Up' }).click();

    // Registration should redirect to the login page
    await expect(page).toHaveURL(/\/login$/);

    // Log in
    await page.locator('#username').fill(username);
    await page.locator('#password').fill(password);
    await page.locator('#login-button').click();

    // Check that the user is logged in
    await expect(page.locator('#nav-logout')).toContainText(username);

    // Log out
    await page.locator('#nav-logout').click();

    // Check that the login link is visible again
    await expect(page.locator('#nav-login')).toBeVisible();
});
