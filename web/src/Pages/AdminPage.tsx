import '../App.css';
import './AdminPage.css';
import { useContext } from 'react';
import { useAdminContext } from '../Hooks/authAdminContext';
import NavigationBar from '../components/NavigationBar';
import AdminListRows from '../components/Admin/AdminList/AdminListRow';
import { UseAdminData } from '../Hooks/useAdminData';
import AdminForm from '../components/Admin/AdminForm';

import { useAdminActions } from '../Hooks/useAdminActions';
import { useNavigate } from 'react-router-dom';
import { AuthContext, IAuthContext } from 'react-oauth2-code-pkce';
import { Tooltip } from 'react-tooltip';
import AuthenticationButtons from '../components/AuthButton';

const AdminPage: React.FC = () => {
  const isAdmin = useAdminContext();

  const { handleAddAdminUser, handleDeleteAdminUser, messageState, resetFormMessage, clearMessageForRow } = useAdminActions();
  const { token }: IAuthContext = useContext(AuthContext);

  const { useradminData, isLoading, isError } = UseAdminData();
  const navigate = useNavigate();
  const handleNavigation = () => {
    navigate('/');
  };

  if (!token) {
    return (
      <div className="center">
        <div className="info-panel warning">You are unauthorized. Please log in.</div>
        <div className="login-button-container">
          <AuthenticationButtons />
        </div>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="center">
        <h2 className="access_denied_title">Access Denied</h2>
        <p className="info-panel warning">You do not have admin privileges.</p>
        <div className="login-button-container" data-tooltip-id="url-shortner-tooltip" data-tooltip-content={'go to URL shortener'}>
          <button onClick={handleNavigation}>URL shortener</button>
        </div>
        <Tooltip id="url-shortner-tooltip" />
      </div>
    );
  }

  return (
    <>
      <NavigationBar></NavigationBar>
      <div className="list-center">
        <AdminForm onSubmit={handleAddAdminUser} message={messageState.formMessage} clearMessage={resetFormMessage}></AdminForm>

        <h3>Current Admin Users</h3>
        {isLoading && <p>Loading...</p>}
        {isError && <p>An error occurred while fetching admin users.</p>}
        {useradminData && (
          <AdminListRows
            data={useradminData}
            onDelete={handleDeleteAdminUser}
            messagesByRow={messageState.messagesByRow}
            clearMessage={clearMessageForRow}
          />
        )}
      </div>
    </>
  );
};

export default AdminPage;
