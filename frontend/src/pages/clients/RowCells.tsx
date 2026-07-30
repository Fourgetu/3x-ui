import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Popover, Space, Tag, Tooltip } from 'antd';
import {
  DeleteOutlined,
  EditOutlined,
  InfoCircleOutlined,
  QrcodeOutlined,
  RetweetOutlined,
} from '@ant-design/icons';

import { formatInboundLabel } from '@/lib/inbounds/label';
import type { InboundOption } from '@/hooks/useClients';

const ICON_BUTTON_STYLE = { fontSize: 16 } as const;

interface ClientRowActionsProps {
  email: string;
  onShowQr: (email: string) => void;
  onShowInfo: (email: string) => void;
  onResetTraffic: (email: string) => void;
  onEdit: (email: string) => void;
  onDelete: (email: string) => void;
}

// Five Tooltip-wrapped buttons per row, none of which depend on traffic. Left
// inline they re-ran rc-tooltip's alignment machinery for every visible row on
// every traffic push — 125 Tooltips on a 25-row page, five seconds apart.
// Keyed on the email rather than the row object, because a push replaces the row
// object of every client whose counters moved; the page resolves the live row.
export const ClientRowActions = memo(function ClientRowActions({
  email,
  onShowQr,
  onShowInfo,
  onResetTraffic,
  onEdit,
  onDelete,
}: ClientRowActionsProps) {
  const { t } = useTranslation();
  return (
    <Space size={4}>
      <Tooltip title={t('pages.clients.qrCode')}>
        <Button
          size="small"
          type="text"
          style={ICON_BUTTON_STYLE}
          icon={<QrcodeOutlined />}
          aria-label={t('pages.clients.qrCode')}
          onClick={() => onShowQr(email)}
        />
      </Tooltip>
      <Tooltip title={t('pages.clients.clientInfo')}>
        <Button
          size="small"
          type="text"
          style={ICON_BUTTON_STYLE}
          icon={<InfoCircleOutlined />}
          aria-label={t('pages.clients.clientInfo')}
          onClick={() => onShowInfo(email)}
        />
      </Tooltip>
      <Tooltip title={t('pages.inbounds.resetTraffic')}>
        <Button
          size="small"
          type="text"
          style={ICON_BUTTON_STYLE}
          icon={<RetweetOutlined />}
          aria-label={t('pages.inbounds.resetTraffic')}
          onClick={() => onResetTraffic(email)}
        />
      </Tooltip>
      <Tooltip title={t('edit')}>
        <Button
          size="small"
          type="text"
          style={ICON_BUTTON_STYLE}
          icon={<EditOutlined />}
          aria-label={t('edit')}
          onClick={() => onEdit(email)}
        />
      </Tooltip>
      <Tooltip title={t('delete')}>
        <Button
          size="small"
          type="text"
          danger
          style={ICON_BUTTON_STYLE}
          icon={<DeleteOutlined />}
          aria-label={t('delete')}
          onClick={() => onDelete(email)}
        />
      </Tooltip>
    </Space>
  );
});

const CHIP_STYLE = { margin: 2 } as const;
const OVERFLOW_CHIP_STYLE = { margin: 2, cursor: 'pointer' } as const;
const OVERFLOW_LIST_STYLE = {
  display: 'flex',
  flexDirection: 'column' as const,
  gap: 4,
  maxWidth: 280,
  maxHeight: 280,
  overflowY: 'auto' as const,
};

interface ClientInboundChipsProps {
  ids: number[];
  inboundsById: Record<number, InboundOption>;
  protocolColors: Record<string, string>;
  chipLimit: number;
  labelForId?: (id: number) => string;
  titleForId?: (id: number) => string;
  nodeLabelForId?: (id: number) => string;
}

// Attachments never change on a traffic push either, so the same memoisation
// applies: one Tooltip per visible chip plus a Popover for the overflow.
export const ClientInboundChips = memo(function ClientInboundChips({
  ids,
  inboundsById,
  protocolColors,
  chipLimit,
  labelForId,
  titleForId,
  nodeLabelForId,
}: ClientInboundChipsProps) {
  if (ids.length === 0) return <span className="cell-empty">—</span>;

  const label = (id: number) => {
    const ib = inboundsById[id];
    return labelForId?.(id) || formatInboundLabel(ib?.tag, ib?.remark, ib?.port);
  };
  const title = (id: number) => titleForId?.(id) || label(id);
  const nodeLabel = (id: number) => nodeLabelForId?.(id) || '';
  const chipLabel = (id: number) => {
    const node = nodeLabel(id);
    if (!node) return label(id);
    return (
      <>
        <span className="client-inbound-node">{node}</span>
        <span className="client-inbound-name">{label(id)}</span>
      </>
    );
  };
  const chip = (id: number) => {
    const proto = (inboundsById[id]?.protocol || '').toLowerCase();
    const node = nodeLabel(id);
    return (
      <Tooltip key={id} title={title(id)}>
        <Tag
          color={protocolColors[proto] ?? 'default'}
          className={node ? 'client-inbound-chip' : undefined}
          style={node ? undefined : CHIP_STYLE}
        >
          {chipLabel(id)}
        </Tag>
      </Tooltip>
    );
  };

  const visible = ids.slice(0, chipLimit);
  const overflow = ids.slice(chipLimit);
  return (
    <>
      {visible.map(chip)}
      {overflow.length > 0 && (
        <Popover
          trigger="click"
          placement="bottomRight"
          content={<div style={OVERFLOW_LIST_STYLE}>{overflow.map(chip)}</div>}
        >
          <Tag color="default" style={OVERFLOW_CHIP_STYLE}>+{overflow.length}</Tag>
        </Popover>
      )}
    </>
  );
});
