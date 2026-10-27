import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, fireEvent, screen } from '@testing-library/react';

import { HttpUtil, Msg } from '@/utils';
import InboundsPage from '@/pages/inbounds/InboundsPage';
import type { InboundListProps } from '@/pages/inbounds/list/types';
import { DBInbound } from '@/models/dbinbound';
import fixture from './golden/fixtures/inbound-full/vless-tcp-reality.json';
import { renderWithProviders } from './test-utils';

const serverLinks = ['vless://uuid@bound.example.test:35001#limited'];

const { hydrateInbound } = vi.hoisted(() => ({ hydrateInbound: vi.fn() }));
const slim = new DBInbound({
  ...fixture,
  nodeId: 7,
  settings: { ...fixture.settings, clients: [] },
});

vi.mock('@/layouts/AppSidebar', () => ({ default: () => null }));
vi.mock('@/hooks/useWebSocket', () => ({ useWebSocket: () => {} }));
vi.mock('@/api/queries/useNodesQuery', () => ({
  useNodesQuery: () => ({ nodes: [{ id: 7, address: 'node.example.test' }], fetched: true }),
}));
vi.mock('@/api/queries/useHostsQuery', () => ({
  useHostsQuery: () => ({ hosts: [], fetched: true }),
}));
vi.mock('@/pages/inbounds/useInbounds', () => ({
  useInbounds: () => ({
    fetched: true,
    dbInbounds: [slim],
    totals: { up: 0, down: 0 },
    subSettings: { enable: true },
    hydrateInbound,
  }),
}));
vi.mock('@/pages/inbounds/list', () => ({
  InboundList: ({ onRowAction }: InboundListProps) => (
    <button onClick={() => onRowAction({ key: 'qrcode', dbInbound: slim })}>Open inbound QR</button>
  ),
}));
vi.mock('@/pages/inbounds/qr/QrCodeModal', () => ({
  default: ({
    clients,
    nodeAddress,
    serverLinks,
  }: {
    clients: unknown[];
    nodeAddress: string;
    serverLinks: string[];
  }) => (
    <output data-testid="qr-data">{JSON.stringify({ clients, nodeAddress, serverLinks })}</output>
  ),
}));

beforeEach(() => {
  vi.mocked(HttpUtil.get).mockResolvedValue(new Msg(true, '', serverLinks));
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('inbound QR detail loading', () => {
  it('loads client credentials before opening the QR from the slim list', async () => {
    hydrateInbound.mockResolvedValue(new DBInbound({ ...fixture, nodeId: 7 }));
    renderWithProviders(<InboundsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'Open inbound QR' }));
    const payload = JSON.parse((await screen.findByTestId('qr-data')).textContent ?? '{}');
    expect(hydrateInbound).toHaveBeenCalledWith(fixture.id);
    expect(payload.clients[0].id).toBe(fixture.settings.clients[0].id);
    expect(payload.nodeAddress).toBe('node.example.test');
    expect(payload.serverLinks).toEqual(serverLinks);
    expect(HttpUtil.get).toHaveBeenCalledWith(
      `/panel/api/inbounds/allLinks?inboundId=${fixture.id}`,
    );
  });

  it('does not fall back to local links when the official renderer fails', async () => {
    hydrateInbound.mockResolvedValue(new DBInbound(fixture));
    vi.mocked(HttpUtil.get).mockResolvedValue(new Msg(false, 'failed', null));
    renderWithProviders(<InboundsPage />);
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Open inbound QR' }));
    });
    expect(screen.queryByTestId('qr-data')).toBeNull();
  });

  it('does not generate a credential-less QR when detail loading fails', async () => {
    hydrateInbound.mockResolvedValue(null);
    renderWithProviders(<InboundsPage />);
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Open inbound QR' }));
    });
    expect(hydrateInbound).toHaveBeenCalledWith(fixture.id);
    expect(screen.queryByTestId('qr-data')).toBeNull();
  });
});
