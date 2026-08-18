import { toast as sonnerToast } from "sonner";
import { apiErrorMessage, type ApiResult } from "@/lib/api";

export function toastError(message: string, title?: string) {
  if (title) sonnerToast.error(title, { description: message });
  else sonnerToast.error(message);
}

export function toastSuccess(message: string, title?: string) {
  if (title) sonnerToast.success(title, { description: message });
  else sonnerToast.success(message);
}

export function toastInfo(message: string, title?: string) {
  if (title) sonnerToast.message(title, { description: message });
  else sonnerToast.message(message);
}

export function toastApiError(result: ApiResult, title?: string) {
  toastError(apiErrorMessage(result), title);
}

export { sonnerToast as toast };
