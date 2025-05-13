// Create URL and Body for request admin/user

import useSWR, { mutate } from 'swr';
import { getAdminUsers, AddAdminUser} from '../Service/Api/AdminService';


const DEFAULT_REFRESH_INTERVAL = 1000 * 60 * 5; // 5 min



const fetcher = async (): Promise<string[]> => {
    try {
        const users = await getAdminUsers();
        return users;
    
    } catch (error) {
        console.error("Error during fetching data in fetcher", error);
        return [];
    }
};
    

export function UseAdminData() {
    const { data: admins, error, isLoading } = useSWR('AdminUser', fetcher, { refreshInterval: DEFAULT_REFRESH_INTERVAL });

    return {
        useradminData: admins || [],
        isLoading,
        isError: error
    };
}

export async function AddAndUpdate(email: string){
    await AddAdminUser(email);

    mutate('AdminUser');
}