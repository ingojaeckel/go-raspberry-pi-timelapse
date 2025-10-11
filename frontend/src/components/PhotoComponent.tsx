import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { DataGrid, GridColDef, GridRowId, GridRenderCellParams, GridRowSelectionModel } from '@mui/x-data-grid';
import { Button, ButtonGroup, Dialog, DialogActions, DialogContent, DialogContentText, Alert, Box } from '@mui/material';
import { Download } from '@mui/icons-material';
import { BaseUrl } from '../conf/config'
import { PhotosResponse } from '../models/response';

export interface PhotosRowData {
  ShowDeletionDialog: boolean,
  Photos: any[],
  Selected: GridRowSelectionModel,
  SelectedFilesParameter: string,
  Error: string,
}

const columns: GridColDef[] = [
  { field: 'fileName', headerName: 'Name', width: 300, renderCell: (p: GridRenderCellParams) => (<a href={BaseUrl + "/file/" + p.value}>{p.value}</a> ) },
  { field: 'fileCreateTime', headerName: 'Created At', width: 300 },
  { field: 'fileSizeBytes', headerName: 'Size', width: 100 },
];

export default function PhotosComponent() {
  const [state, setState] = useState<PhotosRowData>({
    ShowDeletionDialog: false,
    Photos: [],
    Selected: { type: 'include', ids: new Set<GridRowId>() },
    SelectedFilesParameter: "",
    Error: "",
  });

const getPhotos = () => {
  console.log("BaseUrl: ", BaseUrl)
  axios
    .get<PhotosResponse>(BaseUrl + "/photos")
    .then(resp => {
      // After receiving a response, map the PhotosResponse to RowData[] which can be displayed in the data grid.
      if (resp.data && resp.data.Photos) {
        var rows: any[] = [];

        for (var i=0; i<resp.data.Photos.length; i++) {
          let photo = resp.data.Photos[i];
          rows.push({
            id: photo.Name,
            isSelected: false,
            fileName: photo.Name,
            fileCreateTime: photo.ModTime,
            fileSizeBytes: photo.Size})
        }

        setState({
          ShowDeletionDialog: false,
          Photos: rows,
          Selected: { type: 'include', ids: new Set<GridRowId>() },
          SelectedFilesParameter: "",
          Error: "",
        });
      }
    })
    .catch(err => {
      if (err.code === 'ERR_NETWORK' || err.message.includes('Network Error')) {
        setState({
          ...state,
          Error: "Unable to connect to server. Please ensure the server is running.",
        });
      } else {
        setState({
          ...state,
          Error: "Failed to fetch photos: " + (err.message || "Unknown error"),
        });
      }
    });
  }

  useEffect(() => {
    getPhotos()
  }, []);

  const handleSelectionModelChanged = (selectionModel: GridRowSelectionModel) => {
    console.log("selection changed: ", selectionModel);

    var link = "";
    selectionModel.ids.forEach(selected => {
      let selectedPhoto = state.Photos.find(e => e.id === selected);
      if (selectedPhoto) {
        let selectedPhotoFilename = selectedPhoto.fileName;
        link += "f=" + selectedPhotoFilename + "&";
      }
    });

    setState({
      ShowDeletionDialog: false,
      Photos: state.Photos,
      Selected: selectionModel,
      SelectedFilesParameter: link,
      Error: state.Error,
    });
  };

  const handleRefreshClicked = () => getPhotos();
  const deletePhotosClicked = () => {
    if (state.Selected.ids.size > 0) {
      setState({
        ShowDeletionDialog: true,
        Photos: state.Photos,
        Selected: state.Selected,
        SelectedFilesParameter: state.SelectedFilesParameter,
        Error: state.Error,
      })
    }
  };
  const handleDeletionCancelled = () => {
    setState({
      ShowDeletionDialog: false,
      Photos: state.Photos,
      Selected: state.Selected,
      SelectedFilesParameter: state.SelectedFilesParameter,
      Error: state.Error,
    })
  };
  const handleDeletionConfirmed = () => {
    axios
    .get(BaseUrl + "/file/delete?" + state.SelectedFilesParameter)
    .then(_resp => {
      setState({
        ShowDeletionDialog: false,
        Photos: state.Photos,
        Selected: state.Selected,
        SelectedFilesParameter: state.SelectedFilesParameter,
        Error: state.Error,
      });
      getPhotos();
    })
    .catch(err => {
      setState({
        ...state,
        ShowDeletionDialog: false,
        Error: "Failed to delete photos: " + (err.message || "Unknown error"),
      });
    });
  };

  const handleDownload = (url: string) => {
    window.location.href = url;
  };

  return (
    <React.Fragment>
      {state.Error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setState({ ...state, Error: "" })}>
          {state.Error}
        </Alert>
      )}
      <Box sx={{ mb: 2 }}>
        <ButtonGroup color="primary" aria-label="outlined primary button group" sx={{ mb: 2 }}>
          <Button onClick={handleRefreshClicked}>Refresh</Button>
          <Button onClick={deletePhotosClicked}>Delete selected ({state.Selected.ids.size})</Button>
        </ButtonGroup>
        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          <Button 
            variant="outlined" 
            startIcon={<Download />}
            onClick={() => handleDownload(BaseUrl + "/archive/zip")}
          >
            Download all (zip)
          </Button>
          <Button 
            variant="outlined" 
            startIcon={<Download />}
            onClick={() => handleDownload(BaseUrl + "/archive/tar")}
          >
            Download all (tar)
          </Button>
          <Button 
            variant="outlined" 
            startIcon={<Download />}
            onClick={() => handleDownload(BaseUrl + "/archive/zip?" + state.SelectedFilesParameter)}
            disabled={state.Selected.ids.size === 0}
          >
            Download selected ({state.Selected.ids.size})
          </Button>
        </Box>
      </Box>
      <div style={{ height: 500, width: '100%' }}>
        <DataGrid
          rows={state.Photos}
          columns={columns}
          checkboxSelection={true}
          disableRowSelectionOnClick={true}
          onRowSelectionModelChange={handleSelectionModelChanged} />
      </div>
      <Dialog open={state.ShowDeletionDialog} onClose={handleDeletionCancelled} aria-describedby="alert-dialog-description">
        <DialogContent>
          <DialogContentText id="alert-dialog-description">Are you sure you want to delete the selected {state.Selected.ids.size} files?</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeletionCancelled} color="primary">Cancel</Button>
          <Button onClick={handleDeletionConfirmed} color="primary" autoFocus>Delete</Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  );
}
