import { useEffect } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import { Tooltip } from 'react-tooltip';

import './UrlForm.css';
import { UrlData } from '../../../data/Types.ts';
import { isValidUrl } from '../../UrlValidator.ts';

interface FormProps {
  onSubmit: (data: UrlData) => void;
  message: { type: 'success' | 'error'; message: string | null };
  clearMessage: () => void;
}

export function UrlForm({ onSubmit, message, clearMessage }: FormProps) {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<UrlData>({
    defaultValues: {
      path: '',
      url: '',
    },
  });

  // Cleanup (success and error)
  useEffect(() => {
    if (message.message) {
      const timer = setTimeout(() => clearMessage(), 3000);
      return () => clearTimeout(timer);
    }
  }, [message, clearMessage]);

  const onError = () => {
    console.log('Wrong');
  };

  // Validation before sending data to app.tsx
  const onSubmitHandler: SubmitHandler<UrlData> = async (data) => {
    try {
      const formData = {
        path: (data.path = data.path.toLocaleLowerCase()),
        url: (data.url = data.url.toLocaleLowerCase()),
      };

      const regex = /^https?:\/\//i;

      if (!regex.test(data.url)) {
        data.url = `https://${data.url}`;
      }

      if (!isValidUrl(data.url)) {
        console.log('url is not valid');
        setError('url', { type: 'error', message: 'The provided URL is not valid. Please try again.' });
      } else {
        console.log('Validation: OK');
        await onSubmit(data);
        reset();
      }
    } catch (err) {
      console.error(err);
      message.type = 'error';
      message.message = 'An error occurred while submitting. Please try again.';
    }
  };

  return (
    <div className="url-generator-container">
      <p className="infoText">Enter the url you want to shorten</p>
      <form onSubmit={handleSubmit(onSubmitHandler, onError)}>
        <div className="inline-container">
          <p>k.nhn.no/</p>
          <input
            aria-label="Enter path"
            placeholder=" Path"
            {...register('path', {
              required: 'Path is required',
            })}
            className={`inputPath ${errors.path ? 'error-border' : ''}`}
            data-tooltip-id="path-tooltip"
            data-tooltip-content={'Name your path'}
          />
          <i className="pi pi-angle-double-right pil" />
          <input
            placeholder=" Url"
            aria-label="Enter url"
            {...register('url', {
              required: 'Long url is required.',
            })}
            className={`${errors.url ? 'error-border' : ''}`}
            data-tooltip-id="url-tooltip"
            data-tooltip-content={'Enter the url you want to shorten'}
          />
          <button disabled={isSubmitting} data-tooltip-id="add-tooltip" data-tooltip-content={'Add url shortener'}>
            {isSubmitting ? <i className="pi pi-spinner-dotted" /> : <i className="pi pi-save" />}
          </button>
          <button
            type="button"
            onClick={() => {
              reset();
            }}
            data-tooltip-id="clear-tooltip"
            data-tooltip-content="Clear inputs"
          >
            <i className="pi pi-eraser"></i>
          </button>
        </div>
        <div className="clearField-container">
          {errors.path && <p className="warning">{errors.path.message}</p>}
          {errors.url && <p className="warning">{errors.url.message}</p>}
          {message.message && <p className={message.type === 'success' ? 'success' : 'warning'}>{message.message}</p>}
        </div>
        <div></div>
      </form>
      <Tooltip id="clear-tooltip" />
      <Tooltip id="add-tooltip" />
      <Tooltip id="path-tooltip" />
      <Tooltip id="url-tooltip" />
    </div>
  );
}

export default UrlForm;
