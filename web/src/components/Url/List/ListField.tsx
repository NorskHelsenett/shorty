import 'primeicons/primeicons.css';
import { Tooltip } from 'react-tooltip';
import 'react-tooltip/dist/react-tooltip.css';
import { useEffect, useState } from 'react';
import 'primereact/resources/themes/saga-blue/theme.css';
import 'primereact/resources/primereact.min.css';
import { Dropdown } from 'primereact/dropdown';

import ReadOnlyRows from './ReadList/ReadList.tsx';
import EditableRow from './EditItem/EditableRow.tsx';
import { SearchInput } from './SearchField/SearchInput.tsx';
import './List.css';
import './SearchField/SearchInput.css';

import type { QrData, ShortenedURLs, UrlData } from '../../../data/Types.ts';
import QrCodeViewer from './QrCodeViewer/QrCodeViewer.tsx';
import { createPortal } from 'react-dom';
import { SortButtons } from '../SortetButtons.tsx';

interface ListProps {
  urls: ShortenedURLs;
  onItemDelete: (data: UrlData) => void;
  onItemUpdate: (data: UrlData) => void;
  message: { type: 'success' | 'error'; message: string | null };
  clearMessage: () => void;
  sortedColumn: string;
  sortOrder: 'asc' | 'desc';
  setSortedColumn: (column: string) => void;
  setSortOrder: (order: 'asc' | 'desc') => void;
}

export function ListField({
  urls,
  onItemDelete,
  onItemUpdate,
  message,
  clearMessage,
  setSortedColumn,
  sortOrder,
  setSortOrder,
  sortedColumn,
}: ListProps) {
  // Hooks
  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(0);
  const [editIndex, setEditIndex] = useState<number | null>(null); // index null = show row, != null = edit row
  const [showQrCode, setShowQrCode] = useState(false);
  const [qrPath, setQrPath] = useState<QrData>({
    path: '',
    url: '',
  });
  const [messagesByRow, setMessagesByRow] = useState<Record<number, { type: 'success' | 'error'; message: string | null }>>({});

  // Variables
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const filteredUrls = urls.filter((item) => item.path.toLowerCase().includes(search.toLowerCase()));
  const indexOfFirstRow = currentPage * rowsPerPage;
  const indexOfLastRow = (currentPage + 1) * rowsPerPage;
  const currentRows = filteredUrls.slice(indexOfFirstRow, indexOfLastRow);
  const rowsPerPageOptions = [
    { id: '5', name: '5', value: 5 },
    { id: '15', name: '15', value: 15 },
    { id: '25', name: '25', value: 25 },
  ];

  // Functions
  const nextPage = () => {
    if (indexOfLastRow < urls.length) {
      setCurrentPage(currentPage + 1);
    }
  };

  const prevPage = () => {
    if (indexOfFirstRow > 0) {
      setCurrentPage(currentPage - 1);
    }
  };

  const showMessageForRow = (rowIndex: number, type: 'success' | 'error', message: string) => {
    setMessagesByRow((prev) => ({
      ...prev,
      [rowIndex]: { type, message },
    }));

    // Remove message after 3 sec
    setTimeout(() => {
      setMessagesByRow((prev) => {
        const copy = { ...prev };
        delete copy[rowIndex];
        return copy;
      });
    }, 3000);
  };

  const handleDelete = async (path: string, rowIndex: number) => {
    const shouldRemove = confirm('Are you sure you want to delete path ' + path + '?');
    if (shouldRemove) {
      const urlData = urls.find((url) => url.path === path);
      if (urlData) {
        await onItemDelete(urlData);
        showMessageForRow(rowIndex, 'success', `Path "${path}" deleted successfully!`);
      } else {
        console.error('URl data not found for path:', path);
        showMessageForRow(rowIndex, 'error', `Failed to delete path: ${path}.`);
      }
    } else {
      handleCancel;
    }
  };

  // sends updated data to app.tsx
  const handleUpdate = async (data: UrlData) => {
    await onItemUpdate(data);
    setEditIndex(null);
  };

  const handleSearchChange = (value: string) => {
    setSearch(value);
  };

  // show Edit-component
  const handleEdit = (index: number) => {
    setEditIndex(index); // edit row
  };

  const handleCancel = () => {
    setEditIndex(null);
  };

  const handleOnQrCodeClick = (data: QrData) => {
    setQrPath(data);
    setShowQrCode(true);
  };

  const handleCloseQr = () => {
    setShowQrCode(false);
  };

  useEffect(() => {
    if (message.message) {
      const timer = setTimeout(() => clearMessage(), 3000);
      return () => clearTimeout(timer);
    }
  }, [message, clearMessage]);

  useEffect(() => {
    // If we are on a page > 0 and the current page is empty, set currentPage to 0 (page 1)
    if (currentPage > 0 && currentRows.length === 0) {
      setCurrentPage(0);
    }
  }, [currentPage, currentRows]);

  if (Array.isArray(urls) && urls.length === 0) {
    return <p>NO URLS available, write the first one</p>;
  }

  return (
    <>
      <div className="edit-container">
        <div className="search-container">
          <h2>Saved paths</h2>
          <SearchInput search={search} onSearchChange={handleSearchChange} />
        </div>
        <div className="list-box">
          <div className="sorting-buttons">
            <div className="sorting-buttons-left">
              <SortButtons sortOrder={sortOrder} setSortOrder={setSortOrder} sortedColumn={sortedColumn} setSortedColumn={setSortedColumn} />
            </div>
            <div className="center-text">
              <p>Total number of paths: {urls.length}</p>
            </div>
          </div>
          <div>
            {currentRows.map((urls, index) => (
              <div key={urls.path}>
                {editIndex === index ? (
                  <>
                    <EditableRow
                      data={urls}
                      onCancel={handleCancel}
                      onUpdate={async (formData) => {
                        await onItemUpdate(formData);
                        showMessageForRow(index, 'success', 'Path updated successfully!');
                        setEditIndex(null);
                      }}
                      message={messagesByRow[index] ?? undefined}
                      clearMessage={clearMessage}
                    />
                  </>
                ) : (
                  <ReadOnlyRows
                    data={urls}
                    index={index}
                    onEdit={handleEdit}
                    onDelete={() => handleDelete(urls.path, index)}
                    onQrClick={handleOnQrCodeClick}
                  />
                )}
                {messagesByRow[index] && (
                  <p className={messagesByRow[index]?.type === 'success' ? 'info-panel success' : 'info-panel error'}>
                    {messagesByRow[index]?.message}
                  </p>
                )}
              </div>
            ))}
          </div>
          {/* Pagination Buttons */}

          <div className="button-container">
            {urls.length > rowsPerPage ? (
              <button onClick={prevPage} disabled={currentPage === 0} data-tooltip-id="prev-tooltip" data-tooltip-content="Previous">
                <i className="pi pi-arrow-left"></i>
              </button>
            ) : null}
            <div className="dropdown">
              <Dropdown
                value={rowsPerPage}
                onChange={(e) => setRowsPerPage(e.value)}
                options={rowsPerPageOptions}
                optionLabel="name"
                placeholder="Select rows per page"
                className="w-full md:w-14rem"
                checkmark={true}
                highlightOnSelect={false}
              />
            </div>
            {urls.length > rowsPerPage ? (
              <button onClick={nextPage} disabled={indexOfLastRow >= urls.length} data-tooltip-id="next-tooltip" data-tooltip-content="Next">
                <i className="pi pi-arrow-right"></i>
              </button>
            ) : null}
          </div>

          <Tooltip id="edit-tooltip" />
          <Tooltip id="delete-tooltip" />
          <Tooltip id="next-tooltip" />
          <Tooltip id="prev-tooltip" />
          <Tooltip id="save-tooltip" />
          <Tooltip id="cancel-tooltip" />
          <Tooltip id="qrcode-tooltip" />
          <Tooltip id="sortPath-tooltip" />
          <Tooltip id="sortURL-tooltip" />
        </div>
      </div>
      {/* QR Code Viewer */}
      {showQrCode ? createPortal(<QrCodeViewer imagePath={qrPath} closeQr={handleCloseQr} isVisible={showQrCode} />, document.body) : null}
    </>
  );
}
