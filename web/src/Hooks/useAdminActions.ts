import { toast } from "react-toastify";
import { DeleteAdminUser } from "../Service/Api/AdminService";
import { AddAndUpdate } from "./useAdminData";
import { useState } from "react";

interface MessageState {
    formMessage: { type: 'success' | 'error'; message: string | null };
    messagesByRow: ({ type: 'success' | 'error'; message: string | null }| undefined)[];
  }

export function useAdminActions() {
    const [messageState, setMessageState] = useState<MessageState>({
            formMessage: { type: 'success', message: null },
            messagesByRow: [],
          });
        
        const setFormMessage = (type: 'success' | 'error', message: string | null) => {
        setMessageState((prev) => ({
            ...prev,
            formMessage: { type, message },
        }));
        };
    
        const setMessageForRow = (index: number, type: 'success' | 'error', message: string | null) => {
            setMessageState((prev) => {
              const newMessages = [...prev.messagesByRow];
              newMessages[index] = { type, message }; 
              return { ...prev, messagesByRow: newMessages }; 
            });
          };

        const clearMessageForRow = (index: number) => {
        setMessageState((prev) => {
            const newMessages = [...prev.messagesByRow];
            newMessages[index] = undefined; 
            return { ...prev, messagesByRow: newMessages };
        });
        };
    
    const handleAddAdminUser = async (email: string) => {

        try {
            await AddAndUpdate(email);
            toast.success('Admin user added successfully!');
            setFormMessage('success', 'email is added as admin successfully!');

        } catch (error) {
            console.error('Failed to add new admin user:', error);
            let errorMessage = "An unexpected error occurred. Please try again later."
            if ((error as any).status === 409) {
                errorMessage = `Email "${email}" already exists as an admin user.`;
            } else if ((error as any).status === 400) {
                errorMessage = 'Invalid input. Pleas check the data and try again';
            }

            setFormMessage('error', errorMessage)
        }
    };


    const handleDeleteAdminUser = async (email: string, index: number)=> {
        try {
            const shouldRemove = confirm('Are you sure you want to delete admin user: "' + email + '"?');

            if (!shouldRemove) return;
            await DeleteAdminUser(email);
            setMessageForRow(index, 'success', 'email deleted successfully!');
        } catch (err) {
            console.error(err);
            setMessageForRow(index ,'error','Failed to delete admin user.');
        }
    };

    const resetFormMessage = () => setFormMessage('success', null);


    return {
        handleAddAdminUser, 
        handleDeleteAdminUser,
         messageState,
        resetFormMessage,
        setFormMessage,
        setMessageForRow,
        clearMessageForRow,
    };
}