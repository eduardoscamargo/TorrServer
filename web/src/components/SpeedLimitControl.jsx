import { useEffect, useState } from 'react'
import axios from 'axios'
import {
  Box,
  Button,
  CircularProgress,
  DialogActions,
  DialogContent,
  DialogTitle,
  ListItemIcon,
  ListItemText,
  TextField,
  Typography,
} from '@material-ui/core'
import SpeedIcon from '@material-ui/icons/Speed'
import { useTranslation } from 'react-i18next'
import { StyledDialog, StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import { settingsHost } from 'utils/Hosts'
import { isStandaloneApp } from 'utils/Utils'

const presets = [
  { label: '1 MB/s', kb: 1024 },
  { label: '5 MB/s', kb: 5120 },
  { label: '10 MB/s', kb: 10240 },
]

export default function SpeedLimitControl({ isOffline, isLoading }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [status, setStatus] = useState(null)
  const [customKB, setCustomKB] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return undefined
    let active = true
    setBusy(true)
    setError('')
    axios
      .post(settingsHost(), { action: 'speed_get' })
      .then(({ data }) => {
        if (!active) return
        setStatus(data)
        setCustomKB(String(data.configured_kb))
      })
      .catch(() => {
        if (active) setError(t('SpeedLimit.LoadFailed'))
      })
      .finally(() => {
        if (active) setBusy(false)
      })
    return () => {
      active = false
    }
  }, [open, t])

  const updateLimit = async request => {
    setBusy(true)
    setError('')
    try {
      const { data } = await axios.post(settingsHost(), request)
      setStatus(data)
      setCustomKB(String(data.configured_kb))
    } catch (requestError) {
      setError(requestError.response?.data?.error || t('SpeedLimit.SaveFailed'))
    } finally {
      setBusy(false)
    }
  }

  const applyLimit = (kb, temporary = false) => updateLimit({ action: 'speed_set', download_rate_limit: kb, temporary })

  const restoreLimit = () => updateLimit({ action: 'speed_restore' })

  const applyCustom = () => {
    const kb = Number(customKB)
    if (!/^\d+$/.test(customKB) || !Number.isSafeInteger(kb)) {
      setError(t('SpeedLimit.InvalidLimit'))
      return
    }
    applyLimit(kb)
  }

  const formatLimit = kb => (kb === 0 ? t('SpeedLimit.Unlimited') : `${Number((kb / 1024).toFixed(2))} MB/s`)

  return (
    <>
      <StyledMenuButtonWrapper disabled={isOffline || isLoading} button onClick={() => setOpen(true)}>
        {isStandaloneApp ? (
          <>
            <SpeedIcon />
            <div>{t('SpeedLimit.Title')}</div>
          </>
        ) : (
          <>
            <ListItemIcon>
              <SpeedIcon />
            </ListItemIcon>
            <ListItemText primary={t('SpeedLimit.Title')} />
          </>
        )}
      </StyledMenuButtonWrapper>

      <StyledDialog open={open} onClose={() => setOpen(false)} fullWidth maxWidth='sm'>
        <DialogTitle>{t('SpeedLimit.Title')}</DialogTitle>
        <DialogContent>
          {busy && !status ? (
            <CircularProgress size={24} />
          ) : (
            <>
              {status && (
                <Typography gutterBottom>
                  {t('SpeedLimit.Current', { limit: formatLimit(status.effective_kb) })}
                  {status.temporary && ` · ${t('SpeedLimit.Temporary')}`}
                  {status.temporary && ` · ${t('SpeedLimit.Saved', { limit: formatLimit(status.configured_kb) })}`}
                </Typography>
              )}
              <Typography variant='body2' color='textSecondary' gutterBottom>
                {t('SpeedLimit.Note')}
              </Typography>
              <Box display='flex' flexWrap='wrap' style={{ gap: 8, marginTop: 16, marginBottom: 16 }}>
                <Button variant='outlined' disabled={busy} onClick={() => applyLimit(0)}>
                  {t('SpeedLimit.Unlimited')}
                </Button>
                {presets.map(preset => (
                  <Button key={preset.kb} variant='outlined' disabled={busy} onClick={() => applyLimit(preset.kb)}>
                    {preset.label}
                  </Button>
                ))}
              </Box>
              <Box display='flex' alignItems='center' flexWrap='wrap' style={{ gap: 8 }}>
                <TextField
                  type='number'
                  label={t('SpeedLimit.Custom')}
                  value={customKB}
                  onChange={event => setCustomKB(event.target.value)}
                  inputProps={{ min: 0, step: 1 }}
                  helperText={t('SpeedLimit.Unit')}
                  disabled={busy}
                />
                <Button color='secondary' variant='contained' disabled={busy} onClick={applyCustom}>
                  {t('SpeedLimit.Apply')}
                </Button>
              </Box>
              <Box display='flex' flexWrap='wrap' style={{ gap: 8, marginTop: 24 }}>
                <Button disabled={busy || !status} onClick={() => applyLimit(0, true)}>
                  {t('SpeedLimit.TemporaryUnlimited')}
                </Button>
                {status?.temporary && (
                  <Button disabled={busy} onClick={restoreLimit}>
                    {t('SpeedLimit.Restore')}
                  </Button>
                )}
              </Box>
              {error && (
                <Typography color='error' role='alert' style={{ marginTop: 16 }}>
                  {error}
                </Typography>
              )}
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)} color='secondary'>
            {t('Close')}
          </Button>
        </DialogActions>
      </StyledDialog>
    </>
  )
}
