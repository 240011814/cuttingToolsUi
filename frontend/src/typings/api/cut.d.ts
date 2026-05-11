declare namespace Api {
  namespace Cut {
    interface BarResult {
      index: number;
      totalLength: number;
      cuts: number[];
      used: number;
      remaining: number;
    }

    interface BinResult {
      binId: number;
      materialType: string; // 新增：材料类型
      materialWidth: number; // 新增：材料宽度
      materialHeight: number; // 新增：材料高度
      pieces: Piece[];
      utilization: number;
    }

    interface CutRecord {
      id: string;

      type: string;

      request: string;

      response: string;

      createTime: string;

      code: string;

      name: string;
    }

    interface CutRecordSearchParams extends Api.Common.CommonSearchParams {
      name?: string | null;
      type?: string | null;
      startTime?: number | null;
      endTime?: number | null;
    }

    interface Piece {
      label: string;
      x: number;
      y: number;
      w: number;
      h: number;
      rotated: boolean;
    }

    interface BinRequest {
      items: Item[];
      materials: Item[];
      height: number;
      width: number;
      strategy: string;
    }

    interface RecordRequest {
      type: string;
      request: string;
      response: string;
      name: string;
    }

    interface Item {
      label: string;
      width: number;
      height: number;
      quantity?: number;
    }

    interface BarItem {
      label?: string;
      length: number;
      quantity: number;
    }

    interface BarRequest {
      items: number[];
      materials: number[];
      newMaterialLength: number;
      loss: number;
      utilizationWeight: number;
    }
  }
}
