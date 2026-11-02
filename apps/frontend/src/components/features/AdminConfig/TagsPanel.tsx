import { useState } from 'react';
import type { ConfigTag } from '../../../types/config';
import { changeAdminConfig } from '../../../services/api';
import { count } from '../../../utils/config';
import EmptyState from '../../ui/EmptyState';
import { AddForm, ChangeStatus, List, Meta, RenameForm, Row } from './shared';
import { dangerButtonClass, smallButtonClass } from './styles';
import { useChanger } from './useChanger';

interface TagsPanelProps {
  tags: ConfigTag[];
  reload: () => void;
}

export default function TagsPanel({ tags, reload }: TagsPanelProps) {
  const changer = useChanger(reload);
  const [renamingId, setRenamingId] = useState<number | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  return (
    <div>
      <p className="mb-4 text-sm text-text-muted">
        Tags describe events. Deleting a tag removes it from its events, no event is deleted. Scrapers that send a tag the site doesn&apos;t have
        have it skipped.
      </p>
      <AddForm label="New tag" placeholder="Name" changer={changer} onAdd={(name) => changeAdminConfig('POST', 'tags', { name })} />
      <ChangeStatus changer={changer} />

      {tags.length === 0 ? (
        <EmptyState title="No tags">Add the first one above.</EmptyState>
      ) : (
        <List>
          {tags.map((tag) => (
            <Row key={tag.id}>
              {renamingId === tag.id ? (
                <RenameForm
                  label="Tag name"
                  initial={tag.name}
                  changer={changer}
                  onSave={(name) => changeAdminConfig('PATCH', `tags/${tag.id}`, { name })}
                  onDone={() => setRenamingId(null)}
                />
              ) : (
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <span>
                    <span className="text-sm font-medium text-text">{tag.name}</span> <Meta>{count(tag.events, 'event')}</Meta>
                  </span>
                  <span className="flex gap-2">
                    <button type="button" onClick={() => setRenamingId(tag.id)} className={smallButtonClass}>
                      Rename
                    </button>
                    <button type="button" onClick={() => setDeletingId(deletingId === tag.id ? null : tag.id)} className={dangerButtonClass}>
                      Delete
                    </button>
                  </span>
                </div>
              )}
              {deletingId === tag.id && (
                <div className="flex flex-wrap items-center gap-2 rounded-control border border-border bg-bg p-3">
                  <p className="flex-1 text-sm text-text-muted">
                    Remove <span className="font-medium text-text">{tag.name}</span> from {count(tag.events, 'event')} and delete it?
                  </p>
                  <button
                    type="button"
                    disabled={changer.busy}
                    onClick={() => void changer.run(() => changeAdminConfig('DELETE', `tags/${tag.id}`)).then((ok) => ok && setDeletingId(null))}
                    className={dangerButtonClass}
                  >
                    Delete tag
                  </button>
                  <button type="button" onClick={() => setDeletingId(null)} className={smallButtonClass}>
                    Cancel
                  </button>
                </div>
              )}
            </Row>
          ))}
        </List>
      )}
    </div>
  );
}
