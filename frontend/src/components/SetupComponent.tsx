import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { SettingsResponse } from '../models/response'
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import { ButtonGroup, Button, FormControl, InputLabel, Select, MenuItem, FormHelperText } from '@material-ui/core';

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
  const [_, setState] = useState<SettingsResponse>({
    SecondsBetweenCaptures:  0,
    OffsetWithinHour:        0,
    PhotoResolutionWidth:    1000,
    PhotoResolutionHeight:   2000,
    PreviewResolutionWidth:  640,
    PreviewResolutionHeight: 0,
    RotateBy:                0,
    ResolutionSetting:       0,
    DebugEnabled:            false,
  });

  useEffect(() => {
    axios
      .get<SettingsResponse>(BaseUrl + "/configuration")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
        }
      });
  }, []);

  return (
    <React.Fragment>
      <div>
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Time between captures (minutes)</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper" defaultValue={1}>
            <MenuItem value={1}>1</MenuItem>
            <MenuItem value={10}>10</MenuItem>
            <MenuItem value={20}>20</MenuItem>
            <MenuItem value={30}>30</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl><br />
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Offset before first capture (minutes)</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper" defaultValue={0}>
            <MenuItem value={0}>0</MenuItem>
            <MenuItem value={5}>5</MenuItem>
            <MenuItem value={10}>10</MenuItem>
            <MenuItem value={15}>15</MenuItem>
            <MenuItem value={20}>20</MenuItem>
            <MenuItem value={25}>25</MenuItem>
            <MenuItem value={30}>30</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl><br />
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Resolution</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper" defaultValue={0}>
            <MenuItem value={0}>3280x2464</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl><br />
        <FormControl className={classes.formControl} fullWidth>
          <InputLabel id="demo-simple-select-helper-label">Rotation (degrees)</InputLabel>
          <Select labelId="demo-simple-select-helper-label" id="demo-simple-select-helper" defaultValue={0}>
            <MenuItem value={0}>0</MenuItem>
            <MenuItem value={180}>180</MenuItem>
          </Select>
          <FormHelperText></FormHelperText>
        </FormControl>
      </div>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button>Restart</Button>
        <Button>Shutdown</Button>
        <Button>Delete Photos</Button>
      </ButtonGroup>
    </React.Fragment>
  );
}
