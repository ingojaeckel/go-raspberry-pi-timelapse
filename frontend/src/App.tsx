import React from 'react';
import { createBrowserHistory } from "history";
import { NavLink, Router } from 'react-router-dom';
import Grid from '@material-ui/core/Grid';
import HomeIcon from '@material-ui/icons/Home';
import SettingsIcon from '@material-ui/icons/Settings';
import TimelineIcon from '@material-ui/icons/Timeline';
import { Container, List, Avatar, ListItemAvatar, ListItem, ListItemText, Divider, CssBaseline } from '@material-ui/core';
import Switcher from './Switch';

const customHistory = createBrowserHistory();

function App() {
  return (
    <Router history={customHistory}>
      <React.Fragment>
        <CssBaseline />
      <Container fixed>
        <Grid container>
          <Grid item xs={2}>
            <List>
              <ListItem to="/" component={NavLink}>
                <ListItemAvatar>
                  <Avatar>
                    <HomeIcon />
                  </Avatar>
                </ListItemAvatar>
                <ListItemText primary="Photos" />
              </ListItem>
              <ListItem to="/monitoring" component={NavLink}>
                <ListItemAvatar>
                  <Avatar>
                    <TimelineIcon />
                  </Avatar>
                </ListItemAvatar>
                <ListItemText primary="Monitoring" />
              </ListItem>
              <Divider />
              <ListItem to="/settings" component={NavLink}>
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
        </Grid>
      </Container>
      </React.Fragment>
    </Router>
  );
}

export default App;
