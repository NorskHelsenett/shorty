import { mutate } from "swr";
import { AddUrl, DeleteUrl, PatchUrl } from "../Service/Api/UrlService";
import { UrlData } from "../data/Types";
import { useState } from "react";

interface MessageState {
    formMessage: { type: 'success' | 'error'; message: string | null };
    listMessage: { type: 'success' | 'error'; message: string | null };
  }

export function usePathActions(){
    const [messageState, setMessageState] = useState<MessageState>({
        formMessage: { type: 'success', message: null },
        listMessage: { type: 'success', message: null },
      });
    
    const setFormMessage = (type: 'success' | 'error', message: string | null) => {
    setMessageState((prev) => ({
        ...prev,
        formMessage: { type, message },
    }));
    };

    const setListMessage = (type: 'success' | 'error', message: string | null) => {
    setMessageState((prev) => ({
        ...prev,
        listMessage: { type, message },
    }));
    };
    


/** send New data to API */
    const handleOnFormsubmit = async ({ path, url }: UrlData) => {
    console.log(path, url);
    try {
        await AddUrl(path, url);
        mutate('/api/get-urls');
        setFormMessage('success', 'URL has been shortened successfully!');
        // toast.success('URL have been shortened successfully!');
    } catch (err) {
        console.error(err);
        let errorMessage = "An unexpected error occurred. Please try again later."
        if ((err as any).status === 409) {
            // toast.error("Path already exists as an admin user.");
            errorMessage = `Path "${path}" already exists.`;

        } else if ((err as any).status === 400) {
            // toast.error('Invalid input. Pleas check the data and try again');
            errorMessage = 'Invalid input. Please check the url and try again.';

        }

        setFormMessage('error', errorMessage)
    }
    };

    const handleItemDelete = async ({ path }: UrlData) => {
    console.log('HandleItemDelete:', path);
        try {
            await DeleteUrl(path);
            mutate('/api/get-urls');
            setListMessage('success', 'Path deleted successfully!');
        } catch (err) {
            console.error(err);
            setListMessage('error', 'Failed to delete path. Please try again later.');
        }
    };

    const handleItemUpdate = async ({ path, url }: UrlData) => {
        try {
            console.log('HandleItemUpdate: path', path, ', url:', url);
            await PatchUrl(path, url);
            mutate('/api/get-urls');
            setListMessage('success', 'Path updated successfully!');
        } catch (err) {
            console.error(err);
            setListMessage('error', 'Failed to update path. Please try again later.');
        }
    };

    const resetFormMessage = () => setFormMessage('success', null);
    const resetListMessage = () => setListMessage('success', null);

return { handleOnFormsubmit,
    handleItemDelete,
    handleItemUpdate,
    messageState,
    resetFormMessage,
    resetListMessage,
    setFormMessage,
    setListMessage,};
}
