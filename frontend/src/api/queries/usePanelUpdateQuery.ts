import { useQuery } from '@tanstack/react-query';
import { HttpUtil } from '@/utils';
import { keys } from '@/api/queryKeys';
import type { PanelUpdateInfo } from '@/pages/index/PanelUpdateModal';

const EMPTY: PanelUpdateInfo = {
  currentVersion: '',
  latestVersion: '',
  updateAvailable: false,
};

export function usePanelUpdateQuery() {
  const query = useQuery({
    queryKey: keys.server.panelUpdate(),
    queryFn: async ({ signal }) => {
      const msg = await HttpUtil.get<PanelUpdateInfo>(
        '/panel/api/server/getPanelUpdateInfo',
        undefined,
        { silent: true, timeout: 25_000, signal },
      );
      if (!msg?.success || !msg.obj?.latestVersion) {
        throw new Error(msg?.msg || 'Update check failed');
      }
      return msg.obj;
    },
    staleTime: 60_000,
    refetchInterval: 5 * 60_000,
    refetchIntervalInBackground: false,
    refetchOnWindowFocus: true,
    retry: false,
  });

  return {
    info: query.isError ? EMPTY : (query.data ?? EMPTY),
    checking: query.isFetching,
    error: query.error?.message ?? '',
    refresh: query.refetch,
  };
}
