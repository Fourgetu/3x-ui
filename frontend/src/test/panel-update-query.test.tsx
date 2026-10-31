import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';

import { usePanelUpdateQuery } from '@/api/queries/usePanelUpdateQuery';
import PanelUpdateModal from '@/pages/index/PanelUpdateModal';
import { HttpUtil } from '@/utils';
import { renderWithProviders } from './test-utils';

function UpdateCheck() {
  const query = usePanelUpdateQuery();
  return (
    <PanelUpdateModal
      open
      info={query.info}
      checking={query.checking}
      checkError={query.error}
      onCheck={() => query.refresh()}
      onClose={() => {}}
      onBusy={() => {}}
    />
  );
}

const update = {
  success: true,
  msg: '',
  obj: {
    channel: 'stable',
    currentVersion: '3.11.0-fourgetu.1',
    latestVersion: 'v3.11.0-fourgetu.4',
    updateAvailable: true,
  },
};

afterEach(() => vi.mocked(HttpUtil.get).mockReset());

it('reports the old empty failure response and lets a manual retry discover the update', async () => {
  vi.mocked(HttpUtil.get)
    .mockResolvedValueOnce({ success: false, msg: '', obj: null })
    .mockResolvedValueOnce(update);
  renderWithProviders(<UpdateCheck />);

  await screen.findByText(
    'Unable to check for panel updates. Retry or check the server connection to GitHub.',
  );
  expect(screen.queryByText('Panel is up to date')).toBeNull();
  expect(
    screen.getByRole('button', { name: /Update Panel/ }).getAttribute('disabled'),
  ).not.toBeNull();

  fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByText('v3.11.0-fourgetu.4');
  await waitFor(() =>
    expect(
      screen.getByRole('button', { name: /Update Panel/ }).getAttribute('disabled'),
    ).toBeNull(),
  );
  expect(screen.queryByText('Update check failed')).toBeNull();
  expect(HttpUtil.get).toHaveBeenCalledTimes(2);
});

it('does not present stale success as up to date after a failed refresh', async () => {
  vi.mocked(HttpUtil.get)
    .mockResolvedValueOnce({ ...update, obj: { ...update.obj, updateAvailable: false } })
    .mockResolvedValueOnce({
      success: false,
      msg: 'GitHub API HTTP 403; GitHub release page HTTP 503',
      obj: null,
    });
  renderWithProviders(<UpdateCheck />);
  await waitFor(() => expect(screen.getAllByText('Panel is up to date').length).toBeGreaterThan(0));

  fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByText('GitHub API HTTP 403; GitHub release page HTTP 503');
  expect(screen.queryByText('Panel is up to date')).toBeNull();
  expect(
    screen.getByRole('button', { name: /Update Panel/ }).getAttribute('disabled'),
  ).not.toBeNull();
});
