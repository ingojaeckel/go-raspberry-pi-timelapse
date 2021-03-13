import React from 'react';
import { createBrowserHistory } from "history";
import { NavLink, Router } from 'react-router-dom';
import Grid from '@material-ui/core/Grid';
import HomeIcon from '@material-ui/icons/Home';
import SettingsIcon from '@material-ui/icons/Settings';
import TimelineIcon from '@material-ui/icons/Timeline';
import { Container, List, Avatar, ListItemAvatar, ListItem, ListItemText, Divider, CssBaseline } from '@material-ui/core';
import Switcher from './Switch';
import axios from 'axios';

// Register default request interceptors
axios.interceptors.request.use(request => {
  console.log('Starting Request', JSON.stringify(request, null, 2))
  return request
})

const customHistory = createBrowserHistory();

function App() {
  return (
    <Router history={customHistory}>
      <React.Fragment>
        <CssBaseline />
        <Container fixed>
          <Grid container spacing={2}>
            <Grid item xs={2}>
              <List>
                <ListItem to={`${process.env.PUBLIC_URL}/`} component={NavLink}>
                  <ListItemAvatar>
                    <Avatar>
                      <HomeIcon />
                    </Avatar>
                  </ListItemAvatar>
                  <ListItemText primary="Photos" />
                </ListItem>
                <ListItem to={`${process.env.PUBLIC_URL}/monitoring`} component={NavLink}>
                  <ListItemAvatar>
                    <Avatar>
                      <TimelineIcon />
                    </Avatar>
                  </ListItemAvatar>
                  <ListItemText primary="Monitoring" />
                </ListItem>
                <Divider />
                <ListItem to={`${process.env.PUBLIC_URL}/settings`} component={NavLink}>
                  <ListItemAvatar>
                    <Avatar>
                      <SettingsIcon />
                    </Avatar>
                  </ListItemAvatar>
                  <ListItemText primary="Settings" />
                </ListItem>
              </List>
            </Grid>
            <Grid item xs={10}>
              <Switcher />
            </Grid>
            <Grid item xs={10}>
              <div className="footer">version: <a href={"https://github.com/ingojaeckel/go-raspberry-pi-timelapse/commit/" + process.env.REACT_APP_GIT_SHA}>{process.env.REACT_APP_GIT_SHA}</a></div>
            </Grid>
          </Grid>
        </Container>
      </React.Fragment>
    </Router>
  );
}

export default App;
