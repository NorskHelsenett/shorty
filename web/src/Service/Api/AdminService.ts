import { fetchWithToken } from "../Auth/Fetch.ts";
import { API_URL } from "../config.ts";

const urlAdminUserAddress = `${API_URL}/vr/user`;


export async function getAdminUsers(): Promise<string[]> {
    try {
      const response = await fetchWithToken<string[]>(urlAdminUserAddress);
      if (!response || response.length === 0) {
        if (!response) {
            console.info('!response')
        }
        console.warn('No admin users found.');
        return []; // returns empty array if response is empty
      }
      return response;
    } catch (error) {
      console.error('Error fetching admin users:', error);
      throw error;
    }
  }

export async function AddAdminUser(emailInput: string): Promise<void> {
  
  const body = JSON.stringify({ email: emailInput});


  try {
    await fetchWithToken(urlAdminUserAddress, 'POST', {
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body,
    });
  } catch (error) {
    console.error("Failed to add admin user:", error);
    throw error;
  }
}
  

export async function DeleteAdminUser  (email: string): Promise<void> {
const url = urlAdminUserAddress + "/" + email;

try {
    await fetchWithToken(url, 'DELETE', {
    headers: {
        'Accept': 'application/json',
    },
    });
    console.log("Data deleted successfully")
} catch (error) {
    console.error("Failed to delete Admin user:", error);
    throw error;
}
}