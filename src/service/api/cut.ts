import { cutRequest } from '../request';

export function CutBar(data: Api.Cut.BarRequest) {
  return cutRequest<Api.Cut.BarResult[]>({
    url: '/cut/bar',
    method: 'post',
    data
  });
}

export function CutBin(data: Api.Cut.BinRequest) {
  return cutRequest<Api.Cut.BinResult[]>({
    url: '/cut/plane',
    method: 'post',
    data
  });
}
