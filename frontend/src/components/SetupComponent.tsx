import React, { useState, useEffect, ChangeEvent } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { Button, ButtonGroup, Grid, MenuItem, Select, TextField, Typography } from '@mui/material';

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

  useEffect(() => {
    console.log("Getting previous configuration...");
    axios
      .get<SettingsResponse>(BaseUrl + "/configuration")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
          updateForm(resp.data);
        }
      });
  }, []);

  const handleRestartClicked = () => {
    console.log("restart clicked");
    axios.get(BaseUrl + "/admin/restart").then(() => console.log("restart initiated"));
  };

  const handleShutdownClicked = () => {
    console.log("shutdown clicked");
    axios.get(BaseUrl + "/admin/shutdown").then(() => console.log("shutdown initiated"));
  };

  function handleSaveSettingsClicked() {
    axios
      .post<SettingsResponse>(BaseUrl + "/configuration", state)
      .then(resp => {
        console.log("Settings updated to ", resp.data);
        setState(resp.data);
      });
  };

  // TODO replace this with proper binding
  function updateForm(data: SettingsResponse) {
    const fieldSecondsBetweenCaptures = document.getElementById("tfSecondsBetweenCaptures")! as HTMLInputElement
    const fieldOffset = document.getElementById("tfOffset")! as HTMLInputElement
    const fieldRotation = document.getElementById("tfRotation")! as HTMLInputElement
    const fieldQuality = document.getElementById("tfQuality")! as HTMLInputElement

    fieldSecondsBetweenCaptures.value = (data.SecondsBetweenCaptures / 60).toString();
    fieldOffset.value = data.OffsetWithinHour.toString();
    fieldRotation.value = data.RotateBy.toString();
    fieldQuality.value = data.Quality.toString();
  }

  const handleOffsetChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { OffsetWithinHour: parseInt(e.target.value) as number }));
  const handleQualityChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { Quality: parseInt(e.target.value) as number }));
  const handleRotationChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { RotateBy: parseInt(e.target.value) as number }));
  const handleTimeBetweenCapturesChanged = (e: ChangeEvent<HTMLInputElement>) => setState(Object.assign(state, { SecondsBetweenCaptures: (parseInt(e.target.value) as number) * 60 }));

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item xs={6}>
          <Typography component={'span'} gutterBottom>Time between captures (minutes):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfSecondsBetweenCaptures" type="number" onChange={handleTimeBetweenCapturesChanged} defaultValue={state.SecondsBetweenCaptures/60} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }} />
        </Grid>
        <Grid item xs={6}>
          <Typography component={'span'} gutterBottom>Delay within the hour before the first capture or -1 to disable (minutes):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfOffset" type="number" defaultValue={state.OffsetWithinHour} onChange={handleOffsetChanged} inputProps={{ inputMode: 'numeric', pattern: '-?[0-9]+' }} />
        </Grid>
        <Grid item xs={6}>
          <Typography component={'span'} gutterBottom>Photo Resolution (pixels)</Typography>
        </Grid>
        <Grid item xs={6}>
          <Select defaultValue={0}>
            <MenuItem value={0}>3280x2464</MenuItem>
          </Select>
        </Grid>
        <Grid item xs={6}>
          <Typography component={'span'} gutterBottom>Rotation (degrees):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfRotation" type="number" defaultValue={state.RotateBy} onChange={handleRotationChanged} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }} />
        </Grid>
        <Grid item xs={6}>
          <Typography component={'span'} gutterBottom>Photo Quality (0..100%):</Typography>
        </Grid>
        <Grid item xs={6}>
          <TextField id="tfQuality" type="number" defaultValue={state.Quality} onChange={handleQualityChanged} inputProps={{ inputMode: 'numeric', pattern: '[0-9]+' }} />
        </Grid>
        <Grid item xs={6}>
        </Grid>
      </Grid>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleSaveSettingsClicked}>Save</Button>
        <Button onClick={handleRestartClicked}>Restart</Button>
        <Button onClick={handleShutdownClicked}>Shutdown</Button>
      </ButtonGroup>
      <Grid container spacing={2}>
        <Grid item xs={6} />
        <Grid item xs={6}>
          Version: {process.env.REACT_APP_GIT_SHA && (<span>
              <a href={"https://github.com/ingojaeckel/go-raspberry-pi-timelapse/commit/" + process.env.REACT_APP_GIT_SHA}>{process.env.REACT_APP_GIT_SHA_ABBREV}</a> built {process.env.REACT_APP_COMMIT_TIME}
            </span>)}
        </Grid>
      </Grid>
    </React.Fragment>
  );
}
