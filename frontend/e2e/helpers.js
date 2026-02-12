/**
 * Log in with the given credentials.
 */
export async function login(page, username, password) {
  await page.goto('/login');
  await page.getByPlaceholder('Username').fill(username);
  await page.getByPlaceholder('Password').fill(password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  await page.waitForURL('/');
}

/**
 * Register a new user account.
 */
export async function registerUser(page, { username, email, password, displayName }) {
  await page.goto('/register');
  await page.getByPlaceholder('Email').fill(email);
  await page.getByPlaceholder('Username').fill(username);
  if (displayName) {
    await page.getByPlaceholder('Display Name (optional)').fill(displayName);
  }
  await page.getByPlaceholder('Password').fill(password);
  await page.getByRole('button', { name: 'Register' }).click();
}
