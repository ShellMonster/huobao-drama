const WORKSPACE_KEY = 'workspace:active'
const WORKSPACE_DRAMA_KEY = 'workspace:drama_id'

export const enableWorkspace = (dramaId?: string) => {
  sessionStorage.setItem(WORKSPACE_KEY, '1')
  if (dramaId) {
    sessionStorage.setItem(WORKSPACE_DRAMA_KEY, dramaId)
  }
}

export const disableWorkspace = () => {
  sessionStorage.removeItem(WORKSPACE_KEY)
  sessionStorage.removeItem(WORKSPACE_DRAMA_KEY)
}

export const isWorkspaceActive = () => {
  return sessionStorage.getItem(WORKSPACE_KEY) === '1'
}

export const getWorkspaceDramaId = () => {
  return sessionStorage.getItem(WORKSPACE_DRAMA_KEY)
}
