import React, { useEffect } from 'react';
import { useAdminContext } from '../../../Hooks/authAdminContext';
import { Tooltip } from 'react-tooltip';
import './AdminList.css';

interface AdminListProps {
  data: string[];
  onDelete: (email: string, index: number) => void;
  messagesByRow: ({ type: 'success' | 'error'; message: string | null } | undefined)[];
  clearMessage: (index: number) => void;
}

const AdminListRows: React.FC<AdminListProps> = ({ data, onDelete, messagesByRow, clearMessage }) => {
  const isAdmin = useAdminContext();
  const sortedData = [...data].sort((a, b) => a.localeCompare(b));

  useEffect(() => {
    const timers = messagesByRow
      .map((message, index) =>
        message?.message
          ? setTimeout(() => {
              clearMessage(index);
            }, 4000)
          : null,
      )
      .filter((timer) => timer !== null);
    return () => {
      timers.forEach((timer) => clearTimeout(timer!));
    };
  }, [messagesByRow, clearMessage]);

  if (Array.isArray(data) && data.length === 0) {
    return <p>No admin emails available, write the first one</p>;
  }

  return (
    <>
      <div className="list-email-box">
        {sortedData.map((email, index) => (
          <div key={index}>
            <div className="list-item-admin-emails ">
              <div className="list-item-email">{email}</div>
              <div className="list-item-email-actions">
                <button
                  data-tooltip-id="delete-tooltip"
                  data-tooltip-content="Delete admin user"
                  onClick={() => onDelete(email, index)}
                  disabled={!isAdmin}
                >
                  <i className={`pi ${isAdmin ? 'pi-trash' : 'pi-ban'}`}></i>
                </button>
              </div>
            </div>
            {messagesByRow[index] && (
              <p className={messagesByRow[index]?.type === 'success' ? 'info-panel success' : 'info-panel error'}>{messagesByRow[index]?.message}</p>
            )}

            <Tooltip id="delete-tooltip" />
          </div>
        ))}
      </div>
    </>
  );
};

export default AdminListRows;
