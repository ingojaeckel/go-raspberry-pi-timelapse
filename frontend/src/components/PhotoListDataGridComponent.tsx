import React from 'react';
import { DataGrid, RowsProp, ColDef, RowData } from '@material-ui/data-grid';

let rows: RowData[] = [];

const columns: ColDef[] = [
  { field: 'fileName', headerName: 'Name', width: 300 },
  { field: 'fileCreateTime', headerName: 'Created At', width: 200 },
  { field: 'fileSizeBytes', headerName: 'Size', width: 100 },
];

export default function PhotoListDataGridComponent() {
    for (var i=0; i<100; i++) {
        rows.push({ id: (i+1), fileName: 'CIMG'+(i+1)+'.jpg', fileCreateTime: 1, fileSizeBytes: 1024 })
    }
    return (
    <div style={{ height: 500, width: '100%' }}>
      <DataGrid rows={rows} columns={columns} />
    </div>
  );
}