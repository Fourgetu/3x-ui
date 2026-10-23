import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import dayjs from 'dayjs';

import ClientFormModal from '@/pages/clients/ClientFormModal';
import type { ClientRecord } from '@/hooks/useClients';
import { renderWithProviders } from './test-utils';

const NOW = dayjs('2026-10-23T12:34:56').valueOf();

function renderClient(expiryTime: number) {
  vi.spyOn(Date, 'now').mockReturnValue(NOW);
  const save = vi.fn().mockResolvedValue({ success: true });
  const onOpenChange = vi.fn();
  renderWithProviders(
    <ClientFormModal
      open
      mode="edit"
      client={
        {
          email: 'expiry-test',
          uuid: '11111111-1111-1111-1111-111111111111',
          subId: 'expiry-test-sub',
          expiryTime,
          enable: true,
        } as ClientRecord
      }
      inbounds={[]}
      save={save}
      onOpenChange={onOpenChange}
    />,
  );
  return { save, onOpenChange };
}

afterEach(() => vi.mocked(Date.now).mockRestore());

describe('client expiry shortcuts', () => {
  it.each([
    ['+1 month', '2026-12-25T08:09:10'],
    ['+3 months', '2027-02-25T08:09:10'],
    ['+6 months', '2027-05-25T08:09:10'],
    ['+1 year', '2027-11-25T08:09:10'],
  ])('extends an active client with %s only after saving', async (label, expected) => {
    const { save } = renderClient(dayjs('2026-11-25T08:09:10').valueOf());
    fireEvent.click(await screen.findByRole('button', { name: label }));
    expect(save).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue(expected.replace('T', ' '))).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /^save$/i }));
    await waitFor(() => expect(save).toHaveBeenCalledTimes(1));
    expect(save.mock.calls[0][0].expiryTime).toBe(dayjs(expected).valueOf());
  });

  it.each([0, dayjs('2026-09-01T00:00:00').valueOf()])(
    'starts an unset or expired date (%s) from now',
    async (expiry) => {
      renderClient(expiry);
      fireEvent.click(await screen.findByRole('button', { name: '+1 month' }));
      expect(screen.getByDisplayValue('2026-11-23 12:34:56')).toBeTruthy();
    },
  );

  it('clamps month-end dates and accumulates repeated clicks without saving on cancel', async () => {
    const { save, onOpenChange } = renderClient(dayjs('2027-01-31T08:09:10').valueOf());
    const addMonth = await screen.findByRole('button', { name: '+1 month' });
    fireEvent.click(addMonth);
    expect(screen.getByDisplayValue('2027-02-28 08:09:10')).toBeTruthy();
    fireEvent.click(addMonth);
    expect(screen.getByDisplayValue('2027-03-28 08:09:10')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(save).not.toHaveBeenCalled();
  });

  it('clamps leap day when adding a year', async () => {
    renderClient(dayjs('2028-02-29T08:09:10').valueOf());
    fireEvent.click(await screen.findByRole('button', { name: '+1 year' }));
    expect(screen.getByDisplayValue('2029-02-28 08:09:10')).toBeTruthy();
  });

  it('preserves delayed activation duration and hides absolute-date shortcuts', async () => {
    const { save } = renderClient(-30 * 86400000);
    expect(screen.queryByRole('button', { name: '+1 month' })).toBeNull();
    fireEvent.click(await screen.findByRole('button', { name: /^save$/i }));
    await waitFor(() => expect(save).toHaveBeenCalledTimes(1));
    expect(save.mock.calls[0][0].expiryTime).toBe(-30 * 86400000);
  });
});
