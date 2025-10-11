import React, { useState, useEffect, ChangeEvent } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { Button, ButtonGroup, Grid, MenuItem, Select, TextField, Typography, Card, CardContent, Box, Alert, IconButton } from '@mui/material';
import { Edit, Save, Cancel } from '@mui/icons-material';

export default function SetupComponent() {
  const [state, setState] = useState<SettingsResponse>({
    // Initial values shown while current config is loaded async
    SecondsBetweenCaptures:  60,
    OffsetWithinHour:        0,
    PhotoResolutionWidth:    3280,
    PhotoResolutionHeight:   2464,
    PreviewResolutionWidth:  640,
    PreviewResolutionHeight: 480,
    RotateBy:                0,
    ResolutionSetting:       0,
    Quality:                 100,
    DebugEnabled:            false,
  });
  const [editMode, setEditMode] = useState<boolean>(false);
  const [error, setError] = useState<string>("");
  const [validationErrors, setValidationErrors] = useState<{[key: string]: string}>({});

  useEffect(() => {
    console.log("Getting previous configuration...");
    axios
      .get<SettingsResponse>(BaseUrl + "/configuration")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
        }
      })
      .catch(err => {
        if (err.code === 'ERR_NETWORK' || err.message.includes('Network Error')) {
          setError("Unable to connect to server. Please ensure the server is running.");
        } else {
          setError("Failed to fetch configuration: " + (err.message || "Unknown error"));
        }
      });
  }, []);

  const handleRestartClicked = () => {
    console.log("restart clicked");
    axios.get(BaseUrl + "/admin/restart")
      .then(() => console.log("restart initiated"))
      .catch(err => {
        setError("Failed to restart: " + (err.message || "Unknown error"));
      });
  };

  const handleShutdownClicked = () => {
    console.log("shutdown clicked");
    axios.get(BaseUrl + "/admin/shutdown")
      .then(() => console.log("shutdown initiated"))
      .catch(err => {
        setError("Failed to shutdown: " + (err.message || "Unknown error"));
      });
  };

  const validateSettings = (): boolean => {
    const errors: {[key: string]: string} = {};
    
    // Validate time between captures (must be > 0)
    if (state.SecondsBetweenCaptures <= 0) {
      errors.SecondsBetweenCaptures = "Time between captures must be greater than 0";
    }
    
    // Validate rotation (must be 0, 90, 180, or 270)
    if (![0, 90, 180, 270].includes(state.RotateBy)) {
      errors.RotateBy = "Rotation must be 0, 90, 180, or 270 degrees";
    }
    
    // Validate quality (must be between 1 and 100)
    if (state.Quality < 1 || state.Quality > 100) {
      errors.Quality = "Quality must be between 1 and 100";
    }
    
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  function handleSaveSettingsClicked() {
    if (!validateSettings()) {
      setError("Please fix validation errors before saving");
      return;
    }
    
    axios
      .post<SettingsResponse>(BaseUrl + "/configuration", state)
      .then(resp => {
        console.log("Settings updated to ", resp.data);
        setState(resp.data);
        setEditMode(false);
        setError("");
      })
      .catch(err => {
        if (err.code === 'ERR_NETWORK' || err.message.includes('Network Error')) {
          setError("Unable to connect to server. Please ensure the server is running.");
        } else {
          setError("Failed to save configuration: " + (err.message || "Unknown error"));
        }
      });
  };

  const handleEditClicked = () => {
    setEditMode(true);
    setError("");
    setValidationErrors({});
  };

  const handleCancelClicked = () => {
    setEditMode(false);
    setValidationErrors({});
    // Re-fetch configuration to reset any unsaved changes
    axios
      .get<SettingsResponse>(BaseUrl + "/configuration")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
        }
      });
  };

  const handleOffsetChanged = (e: ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value);
    setState({...state, OffsetWithinHour: value });
  };
  
  const handleQualityChanged = (e: ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value);
    setState({...state, Quality: value });
  };
  
  const handleRotationChanged = (e: ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value);
    setState({...state, RotateBy: value });
  };
  
  const handleTimeBetweenCapturesChanged = (e: ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value) * 60;
    setState({...state, SecondsBetweenCaptures: value });
  };

  return (
    <React.Fragment>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError("")}>
          {error}
        </Alert>
      )}
      
      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography variant="h6" component="div">
              Camera Settings
            </Typography>
            {!editMode && (
              <IconButton color="primary" onClick={handleEditClicked}>
                <Edit />
              </IconButton>
            )}
          </Box>
          
          <Grid container spacing={2}>
            <Grid size={6}>
              <Typography component={'span'} gutterBottom>Time between captures (minutes):</Typography>
            </Grid>
            <Grid size={6}>
              {editMode ? (
                <TextField 
                  type="number" 
                  value={state.SecondsBetweenCaptures/60}
                  onChange={handleTimeBetweenCapturesChanged}
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]+', min: 1 }}
                  error={!!validationErrors.SecondsBetweenCaptures}
                  helperText={validationErrors.SecondsBetweenCaptures}
                  fullWidth
                  size="small"
                />
              ) : (
                <Typography component={'span'} color="text.secondary">
                  {state.SecondsBetweenCaptures/60}
                </Typography>
              )}
            </Grid>
            
            <Grid size={6}>
              <Typography component={'span'} gutterBottom>Delay within the hour before the first capture or -1 to disable (minutes):</Typography>
            </Grid>
            <Grid size={6}>
              {editMode ? (
                <TextField 
                  type="number" 
                  value={state.OffsetWithinHour}
                  onChange={handleOffsetChanged}
                  inputProps={{ inputMode: 'numeric', pattern: '-?[0-9]+' }}
                  fullWidth
                  size="small"
                />
              ) : (
                <Typography component={'span'} color="text.secondary">
                  {state.OffsetWithinHour}
                </Typography>
              )}
            </Grid>
            
            <Grid size={6}>
              <Typography component={'span'} gutterBottom>Photo Resolution (pixels)</Typography>
            </Grid>
            <Grid size={6}>
              {editMode ? (
                <Select value={0} fullWidth size="small">
                  <MenuItem value={0}>3280x2464</MenuItem>
                </Select>
              ) : (
                <Typography component={'span'} color="text.secondary">
                  3280x2464
                </Typography>
              )}
            </Grid>
            
            <Grid size={6}>
              <Typography component={'span'} gutterBottom>Rotation (degrees):</Typography>
            </Grid>
            <Grid size={6}>
              {editMode ? (
                <TextField 
                  type="number" 
                  value={state.RotateBy}
                  onChange={handleRotationChanged}
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]+', step: 90 }}
                  error={!!validationErrors.RotateBy}
                  helperText={validationErrors.RotateBy || "Valid values: 0, 90, 180, 270"}
                  fullWidth
                  size="small"
                />
              ) : (
                <Typography component={'span'} color="text.secondary">
                  {state.RotateBy}
                </Typography>
              )}
            </Grid>
            
            <Grid size={6}>
              <Typography component={'span'} gutterBottom>Photo Quality (1-100%):</Typography>
            </Grid>
            <Grid size={6}>
              {editMode ? (
                <TextField 
                  type="number" 
                  value={state.Quality}
                  onChange={handleQualityChanged}
                  inputProps={{ inputMode: 'numeric', pattern: '[0-9]+', min: 1, max: 100 }}
                  error={!!validationErrors.Quality}
                  helperText={validationErrors.Quality}
                  fullWidth
                  size="small"
                />
              ) : (
                <Typography component={'span'} color="text.secondary">
                  {state.Quality}
                </Typography>
              )}
            </Grid>
          </Grid>
          
          {editMode && (
            <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
              <Button 
                variant="contained" 
                startIcon={<Save />}
                onClick={handleSaveSettingsClicked}
              >
                Save
              </Button>
              <Button 
                variant="outlined" 
                startIcon={<Cancel />}
                onClick={handleCancelClicked}
              >
                Cancel
              </Button>
            </Box>
          )}
        </CardContent>
      </Card>
      
      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Typography variant="h6" component="div" sx={{ mb: 2 }}>
            System Actions
          </Typography>
          <ButtonGroup color="primary" aria-label="outlined primary button group">
            <Button onClick={handleRestartClicked}>Restart</Button>
            <Button onClick={handleShutdownClicked}>Shutdown</Button>
          </ButtonGroup>
        </CardContent>
      </Card>
      
      <Grid container spacing={2}>
        <Grid size={12}>
          <Typography variant="body2" color="text.secondary">
            Version: {process.env.REACT_APP_GIT_SHA && (<span>
                <a href={"https://github.com/ingojaeckel/go-raspberry-pi-timelapse/commit/" + process.env.REACT_APP_GIT_SHA}>{process.env.REACT_APP_GIT_SHA_ABBREV}</a> built {process.env.REACT_APP_COMMIT_TIME}
              </span>)}
          </Typography>
        </Grid>
      </Grid>
    </React.Fragment>
  );
}
