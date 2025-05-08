import { useContext } from 'react';
import { AuthContext, IAuthContext } from 'react-oauth2-code-pkce';
import { Tooltip } from 'react-tooltip';

const AuthenticationButtons = () => {
  const auth = useContext(AuthContext);
  const { logOut, logIn }: IAuthContext = useContext(AuthContext);

  const handleLogin = () => {
    logIn();
  };

  const handleLogout = () => {
    logOut();
  };

  return (
    <>
      {auth.token ? (
        <button onClick={handleLogout} data-tooltip-id="logout-tooltip" data-tooltip-content={'logout'}>
          <i className="pi pi-sign-out"></i>
        </button>
      ) : (
        <button onClick={handleLogin} data-tooltip-id="login-tooltip" data-tooltip-content={'login'}>
          <i className="pi pi-sign-in"></i>
        </button>
      )}
      <Tooltip id="logout-tooltip" />
      <Tooltip id="login-tooltip" />
    </>
  );
};

export default AuthenticationButtons;
