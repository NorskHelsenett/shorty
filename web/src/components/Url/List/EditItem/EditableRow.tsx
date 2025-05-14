import { useState } from "react";
import { useForm, SubmitHandler } from "react-hook-form";
import "primeicons/primeicons.css";

import { UrlData } from "../../../../data/Types.ts";
import "./EditableRow.css";
import "../List.css";
import { isValidUrl } from "../../../UrlValidator.ts";

interface EditableRowProps {
  onCancel: () => void;
  data: UrlData;
  onUpdate: (data: UrlData) => void;
  message?: { type: "success" | "error"; message: string | null } | null;
  clearMessage: () => void;
}

const EditableRow: React.FC<EditableRowProps> = ({
  data,
  onCancel,
  onUpdate,
  message,
}) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
    setError: setFormError,
  } = useForm<UrlData>({
    defaultValues: {
      path: data.path,
      url: data.url,
    },
  });

  const [localError, setLocalError] = useState<string | null>(null);

  // Validation before sending data to component
  const onSubmit: SubmitHandler<UrlData> = async (data) => {
    try {
      if (!isValidUrl(data.url)) {
        console.error("url is not valid");
        setLocalError("URL is invalid. Please check the URL and try again.");
        setFormError("url", { message: "URL must start with https://." });
      }
      await onUpdate(data);
      console.error("Validation: OK");
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="list-item">
        <div>{data.path}</div>
        <div>
          <i className="pi pi-angle-double-right arrow" />
        </div>

        <input
          type="url"
          placeholder="Enter long url"
          {...register("url", {
            required: "Url is required",
            pattern: {
              value: /^(https:\/\/)/,
              message: "URL must start with https://",
            },
          })}
          className={`${errors.url ? "error-border" : ""}`}
        />

        <div className="list-item__actions">
          <button
            type="submit"
            data-tooltip-id="save-tooltip"
            data-tooltip-content={"Save changes"}
          >
            Save
          </button>
          <button
            onClick={onCancel}
            data-tooltip-id="cancel-tooltip"
            data-tooltip-content={"Cancel changes"}
          >
            Cancel
          </button>
        </div>
      </div>
      {localError && <p className="warning">{localError}</p>}
      {message && (
        <p className={message.type === "success" ? "success " : "warning"}>
          {message.message}
        </p>
      )}
    </form>
  );
};

export default EditableRow;
