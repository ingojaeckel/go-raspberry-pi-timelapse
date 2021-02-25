import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { DataGrid, ColDef, RowData, SelectionModelChangeParams } from '@material-ui/data-grid';
import { PhotosResponse } from '../models/response';
import { BaseUrl } from '../conf/config'

export interface PhotosRowData {
  Photos: RowData[]
}

const columns: ColDef[] = [
  { field: 'fileName', headerName: 'Name', width: 300 },
  { field: 'fileCreateTime', headerName: 'Created At', width: 300 },
  { field: 'fileSizeBytes', headerName: 'Size', width: 100 },
];

export default function PhotoListDataGridComponent() {
  const [state, setState] = useState<PhotosRowData>({
    Photos: []
  });

  useEffect(() => {
    axios
      .get<PhotosResponse>(BaseUrl + "/photos")
      .then(resp => {
        // After receiving a response, map the PhotosResponse to RowData[] which can be displayed in the data grid.
        if (resp.data) {
          var rows: RowData[] = [];

          for (var i=0; i<resp.data.Photos.length; i++) {
            let photo = resp.data.Photos[i];
            rows.push({id: i, fileName: photo.Name, fileCreateTime: photo.ModTime, fileSizeBytes: photo.Size})
          }

          setState({Photos: rows});
        }
      });
  }, []);


    const handleSelectionModelChanged = (selectionModel: SelectionModelChangeParams) => {
        console.log("selection changed: ", selectionModel);
    };

    return (
    <div style={{ height: 500, width: '100%' }}>
      <DataGrid
        rows={state.Photos}
        columns={columns}
        checkboxSelection={true}
        disableSelectionOnClick={true}
        onSelectionModelChange={handleSelectionModelChanged} />
    </div>
  );
}