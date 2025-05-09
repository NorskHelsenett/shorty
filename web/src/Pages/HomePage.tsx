import "../App.css";
import { useContext, useState } from "react";
import { ToastContainer } from "react-toastify";
import useSWR from "swr";

import Heading from "../components/HeaderField/HeaderField.tsx";
import { ListField } from "../components/Url/List/ListField.tsx";
import { getUrl } from "../Service/Api/UrlService.ts";
import { AuthContext, type IAuthContext } from "react-oauth2-code-pkce";
import NavigationBar from "../components/NavigationBar.tsx";
import { usePathActions } from "../Hooks/usePathActions.ts";
import { UrlForm } from "../components/Url/Form/UrlForm.tsx";
import AuthenticationButtons from "../components/AuthButton.tsx";

const fetcher = async (_url: string) => {
  return getUrl();
};

function App() {
  const { token }: IAuthContext = useContext(AuthContext);
  const {
    handleOnFormsubmit,
    handleItemDelete,
    handleItemUpdate,
    messageState,
    resetFormMessage,
    resetListMessage,
  } = usePathActions();

  /** Get data from api */
  const {
    data: urls,
    isLoading,
    error,
  } = useSWR(token ? "/api/get-urls" : null, fetcher);
  const [sortColumn, setSortColumn] = useState("path");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

  const sortedItems = Array.isArray(urls)
    ? [...urls].sort((a, b) => {
        const res =
          sortColumn === "path"
            ? a.path.localeCompare(b.path)
            : a.url.localeCompare(b.url);
        return sortOrder === "asc" ? res : -res;
      })
    : [];

  if (error && error.status == 401) {
    return <h3>You are unauthorized. Please log in.</h3>;
  }
  if (error) {
    console.log("error start:", error);
    return (
      <div className="center">
        <h3>Something unexpected happened. Try again later</h3>
        <div className="login-button-container">
          <AuthenticationButtons />
        </div>
      </div>
    );
  }
  if (!token) {
    return (
      <>
        <Heading />
        <div className="info-panel warning">
          You are unauthorized. Please log in.
        </div>
        <div className="login-button-container">
          <AuthenticationButtons />
        </div>
      </>
    );
  }

  return (
    <div>
      <NavigationBar></NavigationBar>
      <Heading />
      <UrlForm
        onSubmit={handleOnFormsubmit}
        message={messageState.formMessage}
        clearMessage={resetFormMessage}
      />
      {!isLoading && !Array.isArray(urls) && sortedItems.length === 0 && (
        <div className="center">
          <p className="textCenter border">
            No URLS available, write the first one.
          </p>
        </div>
      )}
      {!isLoading && Array.isArray(urls) && sortedItems.length > 0 ? (
        <ListField
          urls={sortedItems}
          onItemDelete={handleItemDelete}
          onItemUpdate={handleItemUpdate}
          message={messageState.listMessage}
          clearMessage={resetListMessage}
          sortedColumn={sortColumn}
          sortOrder={sortOrder}
          setSortedColumn={setSortColumn}
          setSortOrder={setSortOrder}
        />
      ) : null}
      {isLoading ? <p>Loading…</p> : null}
      <ToastContainer autoClose={2000} position="bottom-right" />
    </div>
  );
}

export default App;
