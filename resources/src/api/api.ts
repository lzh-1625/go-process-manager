import axios, {
  AxiosInstance,
  AxiosError,
  AxiosRequestConfig,
  AxiosResponse,
} from "axios";
import { useSnackbarStore } from "@/stores/snackbarStore";
import router from "../router";

const snackbarStore = useSnackbarStore();

interface Result {
  code: number;
  msg: string;
}

// 请求响应参数，包含data
interface ResultData<T = any> extends Result {
  data?: T;
}
const URL: string = "";
enum RequestEnums {
  TIMEOUT = 20000,
}
const config = {
  // 默认地址
  baseURL: URL as string,
  // 设置超时时间
  timeout: RequestEnums.TIMEOUT as number,
  // 跨域时候允许携带凭证
  withCredentials: true,
};

class RequestHttp {
  service: AxiosInstance;
  public constructor(config: AxiosRequestConfig) {
    // 实例化axios
    this.service = axios.create(config);

    this.service.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem("token") || "";
        config.headers.Authorization = "bearer " + token;
        config.url = "/api" + config.url;
        return config;
      },
      (error: AxiosError) => {
        snackbarStore.showErrorMessage(error);
        return Promise.reject(error);
      },
    );

    /**
     * 响应拦截器
     * 服务器换返回信息 -> [拦截统一处理] -> 客户端JS获取到信息
     */
    this.service.interceptors.response.use(
      (response: AxiosResponse) => {
        const { data } = response; // 解构
        if (data.code < 0) {
          snackbarStore.showErrorMessage(data.message);
          return Promise.reject(data);
        }
        if (data.code > 0) {
          snackbarStore.showSuccessMessage(data.message);
        }
        return data;
      },
      (error: AxiosError) => {
        const { response } = error;
        if (response) {
          this.handleCode(response.status);
        }
        //@ts-ignore
        snackbarStore.showErrorMessage(response.data.message);
      },
    );
  }
  handleCode(code: number): void {
    switch (code) {
      case 401:
        router.replace("/login");
        break;
      default:
        break;
    }
  }

  // 常用方法封装
  get<T>(url: string, params?: object): Promise<ResultData<T>> {
    return this.service.get(url, { params });
  }
  post<T>(url: string, params?: object): Promise<ResultData<T>> {
    return this.service.post(url, params);
  }
  put<T>(url: string, params?: object): Promise<ResultData<T>> {
    return this.service.put(url, params);
  }
  delete<T>(url: string, params?: object): Promise<ResultData<T>> {
    return this.service.delete(url, { params });
  }
}

// 导出一个实例对象
export default new RequestHttp(config);
