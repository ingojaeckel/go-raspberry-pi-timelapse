import React, { useState, useEffect, ChangeEvent } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import { ButtonGroup, Button, FormControl, Select, MenuItem, Slider, Typography } from '@material-ui/core';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    formControl: {
      margin: theme.spacing(1),
      minWidth: 120,
    },
    selectEmpty: {
      marginTop: theme.spacing(2),
    },
  }),
);

export default function SetupComponent() {
  const classes = useStyles();
  const [state, setState] = useState<SettingsResponse>({
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
        }
      });
  }, []);

  const handleRestartClicked = () => {
    console.log("restart clicked");

    axios
      .get(BaseUrl + "/admin/restart")
      .then(() => console.log("restart initiated"));
  };

  const handleShutdownClicked = () => {
    console.log("shutdown clicked");

    axios
      .get(BaseUrl + "/admin/shutdown")
      .then(() => console.log("shutdown initiated"));
  };

  function handleSaveSettingsClicked() {
    axios
      .post<SettingsResponse>(BaseUrl + "/configuration", state)
      .then(resp => {
        console.log("Settings updated to ", resp.data);
        setState(resp.data);
      });
  };

  function handleOffsetChanged(_event: ChangeEvent<{}>, value: number | number[]) {
    setState(Object.assign(state, { OffsetWithinHour: value as number }))
  }
  function handleQualityChanged(_event: ChangeEvent<{}>, value: number | number[]) {
    setState(Object.assign(state, { Quality: value as number }));
  }
  function handleRotationChanged(_event: ChangeEvent<{}>, value: number | number[]) {
    setState(Object.assign(state, { RotateBy: value as number }));
  }
  function handleTimeBetweenCapturesChanged(event: ChangeEvent<{}>, value: number | number[]) {
    const newVal = (value as number) * 60;
    const newState = Object.assign(state, { SecondsBetweenCaptures: newVal })
    setState(newState);
  }

  return (
    <React.Fragment>
      <div>
        <FormControl className={classes.formControl} fullWidth>
          <Typography gutterBottom>Time between captures (minutes): {state.SecondsBetweenCaptures/60}</Typography>
          <Slider valueLabelDisplay="auto" defaultValue={1} step={1} marks min={1} max={30} onChange={handleTimeBetweenCapturesChanged} />
          <Typography gutterBottom>Delay within value hour before first capture (minutes): {state.OffsetWithinHour}</Typography>
          <Slider valueLabelDisplay="auto" defaultValue={0} step={5} marks min={0} max={30} onChange={handleOffsetChanged} />
          <Typography gutterBottom>Photo Resolution (pixels)</Typography>
          <Select defaultValue={0}>
            <MenuItem value={0}>3280x2464</MenuItem>
          </Select>
          <Typography gutterBottom>Rotation (degrees): {state.RotateBy}</Typography>
          <Slider valueLabelDisplay="auto" defaultValue={0} step={180} marks min={0} max={180} onChange={handleRotationChanged} />
          <Typography gutterBottom>Photo Quality (0..100%): {state.Quality}</Typography>
          <Slider valueLabelDisplay="auto" defaultValue={0} step={5} marks min={0} max={100} onChange={handleQualityChanged} />
        </FormControl>
      </div>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={handleSaveSettingsClicked}>Save</Button>
        <Button onClick={handleRestartClicked}>Restart</Button>
        <Button onClick={handleShutdownClicked}>Shutdown</Button>
      </ButtonGroup>
    </React.Fragment>
  );
}
