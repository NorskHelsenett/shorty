import React from 'react';
import { QrData, UrlData } from '../../../../data/Types';
import 'primeicons/primeicons.css';
import './ReadList.css';
import '../List.css';
import { useAdminContext } from '../../../../Hooks/authAdminContext';
interface ReadOnlyRowsProps {
  data: UrlData;
  index: number; // position to elements in array (data)
  onEdit: (index: number) => void;
  onDelete: (path: string) => void;
  onQrClick: (data: QrData) => void;
}

const ReadOnlyRows: React.FC<ReadOnlyRowsProps> = ({ data, index, onEdit, onDelete, onQrClick }) => {
  const isAdmin = useAdminContext();

  var modify = data.modify;
  //console.log('admin:', isAdmin, 'owner:', data.owner, 'modify:', modify, 'data.Modify', data.modify);

  const handleOnClick = () => {
    onQrClick({
      path: data.path,
      url: data.url,
    });
  };

  return (
    <>
      <div className="list-item">
        <div className="list-item-1">{data.path}</div>
        <div>
          <i className="pi pi-angle-double-right arrow" />
        </div>
        <div className="cell-breaweWord">{data.url}</div>
        <div className="list-item__actions">
          <button
            data-tooltip-id="edit-tooltip"
            data-tooltip-content={modify ? 'Edit line' : `owner: ${data.owner}`}
            onClick={() => onEdit(index)}
            disabled={!modify}
          >
            <i className={`pi ${modify ? ' pi-pencil' : 'pi-ban'}`}></i>
          </button>
          <button
            data-tooltip-id="delete-tooltip"
            data-tooltip-content={modify ? 'Delete line' : `owner: ${data.owner}`}
            onClick={() => onDelete(data.path)}
            disabled={!modify}
          >
            <i className={`pi ${modify ? 'pi-trash' : 'pi-ban'}`}></i>
          </button>
          <button data-tooltip-id="qrcode-tooltip" data-tooltip-content="Show QR-code" onClick={handleOnClick}>
            <i className={`pi pi-qrcode`}></i>
          </button>
        </div>
      </div>
    </>
  );
};

export default ReadOnlyRows;
