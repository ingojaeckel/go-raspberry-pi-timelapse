import React from 'react';
import { DataGrid, RowsProp, ColDef } from '@material-ui/data-grid';

const rows: RowsProp = [
  { id: 1, col1: 'Hello', col2: 'World' },
  { id: 2, col1: 'XGrid', col2: 'is Awesome' },
  { id: 3, col1: 'Material-UI', col2: 'is Amazing' },
];

const columns: ColDef[] = [
  { field: 'col1', headerName: 'Column 1', width: 150 },
  { field: 'col2', headerName: 'Column 2', width: 150 },
];

export default function PhotoListDataGridComponent() {
    return (
    <div style={{ height: 300, width: '100%' }}>
      <DataGrid rows={rows} columns={columns} />
    </div>
  );
}