import { request } from '../request';

export function cutBar(data: Api.Cut.BarRequest) {
  return request<Api.Cut.BarResult[]>({
    url: '/cut/bar',
    method: 'post',
    data
  });
}

export function cutBin(data: Api.Cut.BinRequest) {
  return request<Api.Cut.BinResult[]>({
    url: '/cut/plane',
    method: 'post',
    data
  });
}

export function add(data: Api.Cut.RecordRequest) {
  return request<Api.Cut.CutRecord[]>({
    url: '/cutRecord/add',
    method: 'post',
    data
  });
}
