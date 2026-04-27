import { Result } from '../type';
declare const request: <T = Result>(url: string, data?: {
    [key: string]: any;
}) => Promise<T>;
export default request;
