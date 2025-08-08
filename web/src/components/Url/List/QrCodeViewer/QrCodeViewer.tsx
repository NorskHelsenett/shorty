import React from "react";
import { Tooltip } from "react-tooltip";
import "./QrCodeViewer.css";
import { QRCodeCanvas } from "qrcode.react";
import { QrData } from "../../../../data/Types";
import { API_URL } from "../../../../Service/config";

interface QrCodeViewerProps {
  imagePath: QrData;
  closeQr: () => void;
  isVisible: boolean;
}

const qrCodeViewer: React.FC<QrCodeViewerProps> = ({
  imagePath,
  closeQr,
  isVisible,
}) => {
  const handleClose = () => {
    closeQr();
  };

  const downloadQrCode = () => {
    const canvas = document.querySelector(
      "#qrcode-canvas"
    ) as HTMLCanvasElement;
    if (!canvas) throw new Error("<canvas> not found int the DOM");

    const pngUrl = canvas
      .toDataURL("image/png")
      .replace("image/png", "image/octet-stream");
    const downloadLink = document.createElement("a");
    downloadLink.href = pngUrl;
    downloadLink.download = "QR code.png";
    document.body.appendChild(downloadLink);
    downloadLink.click();
    document.body.removeChild(downloadLink);
  };

  const imageUrlPath = API_URL + "/" + imagePath.path;
  const isNhnUrl = imagePath.url.includes("nhn");

  return (
    <>
      <dialog className={"qrcode-dialog"} id="dialog" open={isVisible}>
        <div className="dialog-buttons">
          <button
            id="download"
            data-tooltip-id="downlaod-tooltip"
            data-tooltip-content="Download"
            onClick={downloadQrCode}
          >
            {" "}
            <i className="pi pi-file-o"></i>
          </button>
          <button
            id="close"
            data-tooltip-id="close-tooltip"
            data-tooltip-content="Close"
            onClick={handleClose}
          >
            <i className="pi pi-times"></i>
          </button>
        </div>
        <h4 id="dialog_title">{API_URL + "/" + imagePath.path}</h4>
        <QRCodeCanvas
          value={imageUrlPath}
          id="qrcode-canvas"
          size={200}
          level={"H"}
          imageSettings={
            isNhnUrl
              ? {
                  src: "./public/nhnlogo.png",
                  height: 40,
                  width: 40,
                  excavate: true,
                }
              : undefined
          }
        />
      </dialog>
      <Tooltip id="downlaod-tooltip" />
      <Tooltip id="close-tooltip" />
    </>
  );
};

export default qrCodeViewer;
