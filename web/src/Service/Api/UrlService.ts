
import { UrlData } from '../../data/Types.ts';
import { fetchWithToken } from "../Auth/Fetch.ts";
import { API_URL } from '../config.ts';


const urlAdminAddress = `${API_URL}/admin/`;
const endpoint = urlAdminAddress;

export async function getUrl(): Promise<UrlData[]> {
  return await fetchWithToken<UrlData[]>(urlAdminAddress);
}

export async function AddUrl(pathInput: string, urlInput: string): Promise<void> {
  
  const body = JSON.stringify({ path: pathInput, url: urlInput });

  try {
    await fetchWithToken(endpoint, 'POST', {
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body,
    });
  } catch (error) {
    console.error("Failed to add URL:", error);
    throw error;
  }
}


  export async function PatchUrl(pathInput: string, urlInput: string): Promise<any> {
    const url = urlAdminAddress + pathInput;
    const body = JSON.stringify({ path: pathInput, url: urlInput });

    try{
      const response = await fetchWithToken<any>(url, 'PATCH', {
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body,
      });
      return response;
    } catch (error) {
      console.error('Error patching URL:', error);
      throw error;
    }

  } 

  export async function DeleteUrl(pathInput: string): Promise<void> {
    const url = urlAdminAddress + pathInput;

    try {
      await fetchWithToken(url, 'DELETE', {
        headers: {
          'Accept': 'application/json',
        },
      });
      console.log("Data deleted successfully")
    } catch (error) {
      console.error("Failed to delete URL:", error);
      throw error;
    }
  }