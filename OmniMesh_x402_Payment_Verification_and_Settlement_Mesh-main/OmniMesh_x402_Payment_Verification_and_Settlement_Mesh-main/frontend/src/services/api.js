import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8001/api/v1',
});

export const getContent = async (contentId, receiptId = null) => {
  const url = receiptId ? `/content/${contentId}?receipt_id=${receiptId}` : `/content/${contentId}`;
  const response = await api.get(url);
  return response.data;
};

export const verifyPayment = async (payload) => {
  const response = await api.post('/payments/verify', payload);
  return response.data;
};

export const getReceiptUrl = (receiptId) => {
  return `http://localhost:8001/api/v1/receipts/${receiptId}`;
};
