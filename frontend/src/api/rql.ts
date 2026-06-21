import api from './index'
import type { RQLSearchRequest, RQLSearchResponse } from '../utils/rql/types'

export const rqlApi = {
  search: (data: RQLSearchRequest) => {
    return api.post<RQLSearchResponse>('/rql/search', data)
  }
}
