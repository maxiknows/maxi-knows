
const { test, expect } = require('@playwright/test');

test('homepage loads correctly', async ({ page }) => {
    // Open Maxi Knows
    await page.goto('/');

    // Check the page title
    await expect(page).toHaveTitle('¿Maxi Knows?');

    // Check that the search field is visible
    await expect(page.locator('#search-input')).toBeVisible();

    // Check that the search button is visible
    await expect(page.locator('#search-button')).toBeVisible();
});
