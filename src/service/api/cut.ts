import { request } from '../request';

export function CutBar(data: Api.Cut.BarRequest) {
  return request<Api.Cut.BarResult[]>({
    url: '/cut/bar',
    method: 'post',
    data
  });
}

export function CutBin(data: Api.Cut.BinRequest) {
  return request<Api.Cut.BinResult[]>({
    url: '/cut/plane',
    method: 'post',
    data
  });
}
