// API call
// Transport mechanism for sending request to the API (sends access key to API for validation)


  export async function fetchWithToken<T>(
    url: string, // endpoint
    method: string = 'GET', 
    options?: {
      headers?: HeadersInit;
      body?: BodyInit;
    }
  ): Promise<T> {
    const token = window.localStorage.getItem("ROCP_token")?.replace(/"/g, '');

  
    if (!token) {
      throw new Error("No access token found");
    }
  
    const headers = new Headers(options?.headers);
    headers.append("Authorization", `Bearer ${token}`);
    
  
    try {
      const response = await fetch(url, {
        method, // (GET, POST, DELETE, etc.)
        headers,
        body: options?.body,
      });
  
      if (!response.ok) {
        const error = new Error(`HTTP error ${response.status}: ${response.statusText}`);
        (error as any).status = response.status;
        throw error;
        
      }

      const rawBody = await response.text();

      try {
        return JSON.parse(rawBody) as T; // returns parsed data
      } catch (jsonError) {
        console.error("Failed to parse JSON. Returning raw response body as fallback.", jsonError);
        return rawBody as unknown as T;
      }
       
      
    } catch (error) {
      console.error("Error in fetchWithToken:", error);
      throw error;
    }
  }