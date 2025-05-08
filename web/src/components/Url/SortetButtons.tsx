import { Tooltip } from 'react-tooltip';

interface SortComponent {
  sortedColumn: string;
  sortOrder: 'asc' | 'desc';
  setSortedColumn: (column: string) => void;
  setSortOrder: (order: 'asc' | 'desc') => void;
}

export function SortButtons({ sortOrder, setSortOrder, sortedColumn, setSortedColumn }: SortComponent) {
  const handleSortClick = (column: 'path' | 'url') => {
    if (sortedColumn === column) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortedColumn(column);
      setSortOrder('asc');
    }
  };

  // ascending
  return (
    <>
      <button data-tooltip-id="sortPath-tooltip" data-tooltip-content="Sort path" onClick={() => handleSortClick('path')}>
        {sortedColumn === 'url' && <i className="pi pi-arrows-v"></i>}
        {sortedColumn === 'path' && (sortOrder === 'asc' ? <i className="pi pi-sort-amount-up"></i> : <i className="pi pi-sort-amount-up-alt"></i>)}
      </button>
      <button data-tooltip-id="sortURL-tooltip" data-tooltip-content="Sort URL" onClick={() => handleSortClick('url')}>
        {sortedColumn === 'path' && <i className="pi pi-arrows-v"></i>}
        {sortedColumn === 'url' && (sortOrder === 'asc' ? <i className="pi pi-sort-amount-up"></i> : <i className="pi pi-sort-amount-up-alt"></i>)}
      </button>
      <Tooltip id="sortPath-tooltip" />
      <Tooltip id="sortURL-tooltip" />
    </>
  );
}
