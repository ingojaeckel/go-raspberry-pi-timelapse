import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { DataGrid, GridColDef, GridRowId, GridRenderCellParams, GridRowSelectionModel } from '@mui/x-data-grid';
import { Button, ButtonGroup, Dialog, DialogActions, DialogContent, DialogContentText, Alert, Box, useMediaQuery, useTheme, Link } from '@mui/material';
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

const getColumns = (isMobile: boolean): GridColDef[] => {
  const baseColumns: GridColDef[] = [
    { 
      field: 'fileName', 
      headerName: 'Name', 
      flex: isMobile ? 1 : 0,
      minWidth: isMobile ? 150 : 300,
      renderCell: (p: GridRenderCellParams) => (
        <Link 
          href={BaseUrl + "/file/" + p.value}
          sx={{ 
            color: 'primary.main',
            textDecoration: 'none',
            '&:hover': {
              textDecoration: 'underline',
            },
          }}
        >
          {p.value}
        </Link>
      ) 
    },
    { 
      field: 'fileCreateTime', 
      headerName: 'Created At', 
      flex: isMobile ? 0 : 0,
      minWidth: isMobile ? 130 : 300,
    },
  ];
  
  // Only show file size on larger screens
  if (!isMobile) {
    baseColumns.push({ 
      field: 'fileSizeBytes', 
      headerName: 'Size', 
      width: 100 
    });
  }
  
  return baseColumns;
};

export default function PhotosComponent() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const columns = getColumns(isMobile);
  
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
        <ButtonGroup 
          color="primary" 
          aria-label="outlined primary button group" 
          sx={{ 
            mb: 2,
            flexWrap: { xs: 'wrap', sm: 'nowrap' },
            '& .MuiButtonGroup-grouped': {
              minWidth: { xs: '48%', sm: 'auto' },
              fontSize: { xs: '0.75rem', sm: '0.875rem' },
              padding: { xs: '6px 12px', sm: '6px 16px' },
            }
          }}
        >
          <Button onClick={handleRefreshClicked}>Refresh</Button>
          <Button 
            onClick={deletePhotosClicked}
            aria-label={`Delete selected ${state.Selected.ids.size} photos`}
          >
            Delete ({state.Selected.ids.size})
          </Button>
        </ButtonGroup>
        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          <Button 
            variant="outlined" 
            startIcon={!isMobile && <Download />}
            onClick={() => handleDownload(BaseUrl + "/archive/zip")}
            size={isMobile ? "small" : "medium"}
            aria-label="Download all photos as zip file"
            sx={{ 
              fontSize: { xs: '0.7rem', sm: '0.875rem' },
              padding: { xs: '4px 8px', sm: '6px 16px' },
            }}
          >
            {isMobile ? "Zip" : "Download all (zip)"}
          </Button>
          <Button 
            variant="outlined" 
            startIcon={!isMobile && <Download />}
            onClick={() => handleDownload(BaseUrl + "/archive/tar")}
            size={isMobile ? "small" : "medium"}
            aria-label="Download all photos as tar file"
            sx={{ 
              fontSize: { xs: '0.7rem', sm: '0.875rem' },
              padding: { xs: '4px 8px', sm: '6px 16px' },
            }}
          >
            {isMobile ? "Tar" : "Download all (tar)"}
          </Button>
          <Button 
            variant="outlined" 
            startIcon={!isMobile && <Download />}
            onClick={() => handleDownload(BaseUrl + "/archive/zip?" + state.SelectedFilesParameter)}
            disabled={state.Selected.ids.size === 0}
            size={isMobile ? "small" : "medium"}
            aria-label={`Download selected ${state.Selected.ids.size} photos as zip file`}
            sx={{ 
              fontSize: { xs: '0.7rem', sm: '0.875rem' },
              padding: { xs: '4px 8px', sm: '6px 16px' },
            }}
          >
            {isMobile ? `Selected (${state.Selected.ids.size})` : `Download selected (${state.Selected.ids.size})`}
          </Button>
        </Box>
      </Box>
      <Box sx={{ 
        height: { xs: 'calc(100vh - 320px)', sm: 500 }, 
        width: '100%',
        minHeight: 300,
      }}>
        <DataGrid
          rows={state.Photos}
          columns={columns}
          checkboxSelection={true}
          disableRowSelectionOnClick={true}
          onRowSelectionModelChange={handleSelectionModelChanged}
          pageSizeOptions={[5, 10, 25, 50]}
          initialState={{
            pagination: {
              paginationModel: { pageSize: isMobile ? 10 : 25 },
            },
          }}
          sx={{
            '& .MuiDataGrid-cell': {
              fontSize: { xs: '0.75rem', sm: '0.875rem' },
            },
            '& .MuiDataGrid-columnHeader': {
              fontSize: { xs: '0.75rem', sm: '0.875rem' },
            },
          }}
        />
      </Box>
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
