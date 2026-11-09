
const { test, expect } = require('@playwright/test');

test('user can search for a query', async ({ page }) => {
    // Open Maxi Knows
    await page.goto('/');

    // Enter a search query
    await page.locator('#search-input').fill('JavaScript');

    // Click the search button
    await page.locator('#search-button').click();

    // Check that the URL contains our search query
    await expect(page).toHaveURL(/q=JavaScript/);

    // Check that the search field keeps the entered query
    await expect(page.locator('#search-input')).toHaveValue('JavaScript');
});
