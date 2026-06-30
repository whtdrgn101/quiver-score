import { test, expect } from '@playwright/test';

/*
 * Bracket visualization is a read-only, data-driven route, so we mock the three
 * tournament endpoints it reads (plus /users/me for the auth gate). This keeps
 * the test deterministic and independent of backend tournament seeding.
 */

const CLUB_ID = 'club-1';
const TOURNAMENT_ID = 'tourn-1';
const USER = { id: 'u-evan', username: 'Evan', display_name: 'Evan', email_verified: true };

const ROUNDS = [
  { id: 'r1', round_number: 1, name: 'Quarterfinals', round_type: 'elimination', status: 'completed' },
  { id: 'r2', round_number: 2, name: 'Semifinals', round_type: 'elimination', status: 'completed' },
  { id: 'r3', round_number: 3, name: 'Final', round_type: 'elimination', status: 'completed' },
  // A qualification round that must NOT appear in the bracket.
  { id: 'rq', round_number: 0, name: 'Qualifying', round_type: 'qualification', status: 'completed' },
];

const MATCHUPS = {
  r1: [
    { id: 'm1', match_number: 1, participant_a_id: 'p1', participant_a_name: 'Alice', participant_b_id: 'p2', participant_b_name: 'Bob', score_a: 6, score_b: 4, winner_id: 'p1', winner_name: 'Alice' },
    { id: 'm2', match_number: 2, participant_a_id: 'p3', participant_a_name: 'Cara', participant_b_id: 'p4', participant_b_name: 'Dana', score_a: 2, score_b: 6, winner_id: 'p4', winner_name: 'Dana' },
    { id: 'm3', match_number: 3, participant_a_id: 'p5', participant_a_name: 'Evan', participant_b_id: 'p6', participant_b_name: 'Fay', score_a: 6, score_b: 5, winner_id: 'p5', winner_name: 'Evan' },
    { id: 'm4', match_number: 4, participant_a_id: 'p7', participant_a_name: 'Gus', participant_b_id: null, participant_b_name: null, score_a: null, score_b: null, winner_id: 'p7', winner_name: 'Gus' },
  ],
  r2: [
    { id: 'm5', match_number: 1, participant_a_id: 'p1', participant_a_name: 'Alice', participant_b_id: 'p4', participant_b_name: 'Dana', score_a: 6, score_b: 3, winner_id: 'p1', winner_name: 'Alice' },
    { id: 'm6', match_number: 2, participant_a_id: 'p5', participant_a_name: 'Evan', participant_b_id: 'p7', participant_b_name: 'Gus', score_a: 4, score_b: 6, winner_id: 'p7', winner_name: 'Gus' },
  ],
  r3: [
    { id: 'm7', match_number: 1, participant_a_id: 'p1', participant_a_name: 'Alice', participant_b_id: 'p7', participant_b_name: 'Gus', score_a: 7, score_b: 6, winner_id: 'p1', winner_name: 'Alice' },
  ],
};

const TOURNAMENT = {
  id: TOURNAMENT_ID,
  name: 'Spring Classic',
  status: 'completed',
  organizer_id: 'u-org',
  participants: [],
};

test.describe('Tournament Bracket', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem('access_token', 'test-token');
      localStorage.setItem('refresh_token', 'test-refresh');
    });

    await page.route('**/api/v1/**', (route) => {
      const url = route.request().url();
      const json = (body) =>
        route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) });

      if (url.includes('/users/me')) return json(USER);

      const matchupMatch = url.match(/\/rounds\/([^/]+)\/matchups(?:\?|$)/);
      if (matchupMatch) return json(MATCHUPS[matchupMatch[1]] || []);

      if (/\/tournaments\/[^/]+\/rounds(?:\?|$)/.test(url)) return json(ROUNDS);
      if (/\/tournaments\/[^/]+(?:\?|$)/.test(url)) return json(TOURNAMENT);

      // Layout chrome (sessions, notifications, etc.) — harmless empty default.
      return json([]);
    });
  });

  test('renders the elimination bracket with winners, byes, and champion', async ({ page }) => {
    await page.goto(`/clubs/${CLUB_ID}/tournaments/${TOURNAMENT_ID}/bracket`);

    await expect(page.getByRole('heading', { name: 'Spring Classic' })).toBeVisible();
    await expect(page.getByTestId('bracket')).toBeVisible();

    // 4 + 2 + 1 elimination matchups; the qualification round is excluded.
    await expect(page.getByTestId('bracket-match')).toHaveCount(7);

    // Round labels present.
    await expect(page.getByText('Quarterfinals')).toBeVisible();
    await expect(page.getByText('Semifinals')).toBeVisible();

    // Bye is shown for the empty side of match 4.
    await expect(page.getByText('Bye', { exact: true }).first()).toBeVisible();

    // Current user highlighted.
    await expect(page.getByText('(You)').first()).toBeVisible();

    // Champion callout reflects the decided final.
    await expect(page.getByText('Champion')).toBeVisible();
    await expect(page.getByTestId('bracket-champion')).toContainText('Alice');
  });

  test('shows an empty state when there are no elimination rounds', async ({ page }) => {
    await page.route('**/api/v1/**', (route) => {
      const url = route.request().url();
      const json = (body) =>
        route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) });
      if (url.includes('/users/me')) return json(USER);
      if (/\/tournaments\/[^/]+\/rounds(?:\?|$)/.test(url)) {
        return json([{ id: 'rq', round_number: 1, name: 'Qualifying', round_type: 'qualification', status: 'completed' }]);
      }
      if (/\/tournaments\/[^/]+(?:\?|$)/.test(url)) return json(TOURNAMENT);
      return json([]);
    });

    await page.goto(`/clubs/${CLUB_ID}/tournaments/${TOURNAMENT_ID}/bracket`);
    await expect(page.getByText(/No elimination bracket yet/)).toBeVisible();
  });
});
