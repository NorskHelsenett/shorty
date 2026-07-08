import { Tooltip } from "react-tooltip";
import { useAdminContext } from "../Hooks/useAdminContext";
import { useLocation, useNavigate } from "react-router-dom";
import { useRef } from "react";
import "./NavigationBar.css";
import AuthenticationButtons from "./AuthButton";

const NavigationBar: React.FC = () => {
  const isAdmin = useAdminContext();
  const location = useLocation();
  const navigate = useNavigate();
  const isOnAdminPage = location.pathname === "/user";
  const link = isOnAdminPage ? "/" : "/user";
  const linkText = isOnAdminPage ? "pi pi-home" : "pi pi-key";
  const tooltipText = isOnAdminPage ? "Homepage" : "Adminpage";

  const handleNavigation = () => {
    navigate(link);
  };

  const dialogRef = useRef<HTMLDialogElement | null>(null);

  const handleUserkeyClick = () => {
    const accessKey =
      window.localStorage.getItem("ROCP_token")?.replace(/"/g, "") ||
      "No access key found";
    if (dialogRef.current) {
      const dialogContentElement =
        dialogRef.current?.querySelector(".dialog-content");
      if (dialogContentElement) {
        dialogContentElement.textContent = accessKey;
        dialogRef.current.showModal();
      }
    }
  };

  const handleCloseDialog = () => {
    dialogRef.current?.close();
  };

  const handleCopy = async () => {
    if (!isAdmin) {
      alert("You dont have the permission, you are not admin user!");
      return;
    }
    try {
      const dialogContentElement =
        dialogRef.current?.querySelector(".dialog-content");

      if (!dialogContentElement) {
        throw new Error("Could not find dialog-content");
      }

      const contentToCopy = dialogContentElement.textContent || "";

      if (!contentToCopy.trim()) {
        throw new Error("No content to copy");
      }

      await navigator.clipboard.writeText(contentToCopy);
      alert("Access key copied to clipboard!");
    } catch (error) {
      console.error("Unable to copy to clipboard:", error);
    }
  };

  return (
    <div className="nav-container">
      <div className="nav-bar">
        {isAdmin && isOnAdminPage && (
          <button
            type="button"
            className="nav-button"
            onClick={handleUserkeyClick}
            aria-label="Show access key"
            title="Show access key"
            data-tooltip-id="userkey-button-tooltip"
            data-tooltip-content={"Get your accessKey"}
          >
            <i className={"pi pi-user"}> </i>
          </button>
        )}
        {isAdmin && (
          <button
            type="button"
            className="nav-button"
            onClick={handleNavigation}
            aria-label={tooltipText}
            title={tooltipText}
            data-tooltip-id="nav-button-tooltip"
            data-tooltip-content={tooltipText}
          >
            <i className={linkText}></i>
          </button>
        )}

        <AuthenticationButtons></AuthenticationButtons>
        <dialog className={"nav-userkey-dialog"} ref={dialogRef}>
          <h3 className="nav-userkey-title">
            Your Access Key
          </h3>
          <p className="dialog-content nav-userkey-content"></p>
          <div className="dialog-buttons">
            <button
              type="button"
              onClick={handleCopy}
              aria-label="Copy access key"
              title="Copy access key"
              data-tooltip-id="copy-userkey-button-tooltip"
              data-tooltip-content={"Copy your accessKey"}
            >
              <i className="pi pi-copy" />
            </button>
            <button
              type="button"
              onClick={handleCloseDialog}
              aria-label="Close dialog"
              title="Close dialog"
              data-tooltip-id="close-tooltip"
              data-tooltip-content="Close"
            >
              <i className="pi pi-times"></i>
            </button>
          </div>
        </dialog>
        <Tooltip id="nav-button-tooltip" />
        <Tooltip id="userkey-button-tooltip" />
        <Tooltip id="copy-userkey-button-tooltip" />
        <Tooltip id="close-tooltip" />
        <Tooltip id="log-out-tooltip" />
      </div>
    </div>
  );
};

export default NavigationBar;
