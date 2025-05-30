import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Note Upload Flow', () => {
  test.beforeEach(async ({ page }) => {
    // For now, we'll skip login since it's not implemented yet
    // Just navigate to the dashboard
    await page.goto('http://localhost:3000');
    
    // Wait for the page to be fully loaded
    await page.waitForLoadState('networkidle');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should upload a slide and see updated study stats', async ({ page }) => {
    // Wait for the dashboard to load with a more specific selector
    await page.waitForSelector('h1:has-text("Your Notes")', { 
      state: 'visible',
      timeout: 60000 
    });

    // Click the "Add Note" button using the test ID
    const addNoteButton = page.locator('[data-testid="add-note-button"]');
    await expect(addNoteButton).toBeVisible({ timeout: 60000 });
    await addNoteButton.click();

    // Wait for the upload dialog to appear
    await page.waitForSelector('text=Upload Note', { 
      state: 'visible',
      timeout: 60000 
    });

    // Get the file input using the test ID and upload slide.jpg
    const fileInput = page.locator('[data-testid="file-input"]');
    await expect(fileInput).toBeVisible({ timeout: 60000 });
    await fileInput.setInputFiles(path.join(__dirname, '../fixtures/slide.jpg'));

    // Wait for the upload progress to appear and complete
    await page.waitForSelector('[data-testid="upload-progress"]', { 
      state: 'visible',
      timeout: 60000 
    });

    // Wait for the upload dialog to close
    await page.waitForSelector('text=Upload Note', { 
      state: 'hidden',
      timeout: 60000 
    });

    // Wait for the note to appear in the list
    await page.waitForSelector('[data-testid="notes-list"]', { 
      state: 'visible',
      timeout: 60000 
    });

    // Check if the pie chart shows progress
    const completedPercentage = await page.locator('[data-testid="pie-chart-percentage"]').textContent();
    const value = parseInt(completedPercentage?.replace('%', '') || '0', 10);
    expect(value).toBeGreaterThan(0);

    // Optional: Take a screenshot of the final state
    await page.screenshot({ path: 'test-results/upload-complete.png' });
  });
}); 