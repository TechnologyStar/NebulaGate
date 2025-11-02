import { API } from '../helpers';

export const getAllPackages = async (page, pageSize, status) => {
  const params = new URLSearchParams({
    p: page || 0,
    size: pageSize || 10,
  });
  if (status) {
    params.append('status', status);
  }
  const response = await API.get(`/api/admin/packages?${params.toString()}`);
  return response.data;
};

export const getPackage = async (id) => {
  const response = await API.get(`/api/admin/packages/${id}`);
  return response.data;
};

export const createPackage = async (packageData) => {
  const response = await API.post('/api/admin/packages', packageData);
  return response.data;
};

export const updatePackage = async (id, packageData) => {
  const response = await API.put(`/api/admin/packages/${id}`, packageData);
  return response.data;
};

export const deletePackage = async (id) => {
  const response = await API.delete(`/api/admin/packages/${id}`);
  return response.data;
};

export const getAllRedemptionCodes = async (page, pageSize, packageId, status, code) => {
  const params = new URLSearchParams({
    p: page || 0,
    size: pageSize || 10,
  });
  if (packageId) params.append('package_id', packageId);
  if (status) params.append('status', status);
  if (code) params.append('code', code);
  
  const response = await API.get(`/api/admin/redemption-codes?${params.toString()}`);
  return response.data;
};

export const generateRedemptionCodes = async (packageId, quantity) => {
  const response = await API.post('/api/admin/redemption-codes', {
    package_id: packageId,
    quantity: quantity,
  });
  return response.data;
};

export const revokeRedemptionCode = async (id) => {
  const response = await API.put(`/api/admin/redemption-codes/${id}/revoke`);
  return response.data;
};

export const exportRedemptionCodes = (packageId, status, code) => {
  const params = new URLSearchParams();
  if (packageId) params.append('package_id', packageId);
  if (status) params.append('status', status);
  if (code) params.append('code', code);
  
  window.open(`/api/admin/redemption-codes/export?${params.toString()}`, '_blank');
};

export const redeemPackageCode = async (code) => {
  const response = await API.post('/api/user/redeem', { code });
  return response.data;
};

export const getUserPackages = async () => {
  const response = await API.get('/api/user/packages');
  return response.data;
};

export const getActiveUserPackages = async () => {
  const response = await API.get('/api/user/packages/active');
  return response.data;
};
