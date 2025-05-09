import { fetchWithToken } from "../Auth/Fetch.ts";


const urlAdminUserAddress = "http://localhost:8880/admin/user";


export async function getAdminUsers(): Promise<string[]> {
    try {
      const response = await fetchWithToken<string[]>(urlAdminUserAddress);
      console.log('Fetched admin users:', response);
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
    console.log('Payload being sent:', body);
    console.log('Endpoint being contacted:', urlAdminUserAddress);

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
console.log("deleteAdminUser")

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