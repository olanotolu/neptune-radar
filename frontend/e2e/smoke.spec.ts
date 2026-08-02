import { test, expect } from "@playwright/test";

// Smoke test: the smallest end-to-end check that fails if the app breaks.
// Runs against the Vite dev server (started by playwright.config webServer).
// No backend needed — the token gate and shell render without API calls;
// API views show their empty/loading state, which is what we verify.

// Helper: get past the token gate so the shell renders.
async function enterShell(page: import("@playwright/test").Page) {
  await page.goto("/");
  await page.locator('input[type="password"]').fill("test-token");
  await page.locator('button[type="submit"]').click();
  await expect(page.getByTestId("app-nav")).toBeVisible();
}

test("token gate → shell navigation", async ({ page }) => {
  await page.goto("/");

  // Token gate is visible.
  await expect(page.locator("h2")).toHaveText("Neptune Radar");
  await expect(page.locator('input[type="password"]')).toBeVisible();

  // Enter a dummy token and connect.
  await page.locator('input[type="password"]').fill("test-token");
  await page.locator('button[type="submit"]').click();

  // Shell loaded: header and nav tabs are visible.
  await expect(page.locator(".app-header__title")).toHaveText("Neptune Radar");
  const nav = page.getByTestId("app-nav");
  await expect(nav).toBeVisible();

  // The Today tab is present and active by default.
  const tabs = nav.locator(".app-nav__tab");
  await expect(tabs.first()).toContainText("Today");
  await expect(tabs.filter({ hasText: "Today" })).toHaveClass(/app-nav__tab--active/);
});

test("click through major tabs without crashing", async ({ page }) => {
  await enterShell(page);
  const main = page.getByTestId("app-main");

  // Each major tab should render a non-empty main content area.
  for (const id of ["today", "work", "map", "sources", "feed"]) {
    await page.getByTestId(`nav-${id}`).click();
    // Wait for the tab to become active (navigation happened) then content.
    await expect(page.getByTestId(`nav-${id}`)).toHaveClass(/app-nav__tab--active/);
    await expect(main).not.toBeEmpty();
  }
});

test("search bar navigates to search view", async ({ page }) => {
  await enterShell(page);

  const input = page.getByTestId("app-search-input");
  await input.fill("test query");
  await input.press("Enter");

  // The search tab should now be active.
  await expect(page.getByTestId("nav-search")).toHaveClass(/app-nav__tab--active/);
  // Main content rendered (search view, even if empty results).
  await expect(page.getByTestId("app-main")).not.toBeEmpty();
});

test("dark mode toggle flips the theme attribute", async ({ page }) => {
  await enterShell(page);

  // Default: no dark theme attribute on <html>.
  await expect(page.locator("html")).not.toHaveAttribute("data-theme", "dark");

  const toggle = page.getByTestId("dark-mode-toggle");
  await toggle.click();

  // After click: dark theme is on.
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

  // Click again → back to light.
  await toggle.click();
  await expect(page.locator("html")).not.toHaveAttribute("data-theme", "dark");
});

test("? keyboard shortcut opens the help overlay", async ({ page }) => {
  await enterShell(page);

  // No dialog yet.
  await expect(page.locator('role=dialog[name="Keyboard shortcuts"]')).toHaveCount(0);

  // Press ? to open help.
  await page.keyboard.press("Shift+Slash");

  // Help overlay dialog appears.
  await expect(page.locator('role=dialog[name="Keyboard shortcuts"]')).toBeVisible();

  // Close it with Escape.
  await page.keyboard.press("Escape");
  await expect(page.locator('role=dialog[name="Keyboard shortcuts"]')).toHaveCount(0);
});
