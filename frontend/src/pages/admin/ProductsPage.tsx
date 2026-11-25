import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Button } from '../../components/Button'
import { Card } from '../../components/Card'
import { SelectField, TextField } from '../../components/TextField'
import { fieldErrorsOf, isApiError, useAppState } from '../../lib/app-state'
import { api } from '../../lib/api'
import { mediaUrl } from '../../lib/media'
import type { ApiSupplier } from '../../types/catalog'
import { AdminFrame } from './AdminFrame'

type TierDraft = { minQty: string; unitPrice: string }

const emptyTier = (): TierDraft => ({ minQty: '', unitPrice: '' })

export function AdminProductsPage() {
  return (
    <AdminFrame
      title="Products"
      lede="Add a product for a supplier, with the wholesale tiers buyers see in the catalog."
    >
      <ProductManager />
    </AdminFrame>
  )
}

function ProductManager() {
  const { loadAdminSuppliers, createProduct } = useAppState()
  const [suppliers, setSuppliers] = useState<ApiSupplier[]>([])
  const [supplierId, setSupplierId] = useState('')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState('')
  const [tags, setTags] = useState('')
  const [unit, setUnit] = useState('')
  const [retailPrice, setRetailPrice] = useState('')
  const [photo, setPhoto] = useState<File | null>(null)
  const [preview, setPreview] = useState('')
  const [tiers, setTiers] = useState<TierDraft[]>([emptyTier()])
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  const [problem, setProblem] = useState(false)
  const [pending, setPending] = useState(false)
  const [blocked, setBlocked] = useState('')

  useEffect(() => {
    let cancel = false
    loadAdminSuppliers()
      .then((items) => {
        if (cancel) return
        setSuppliers(items)
        setSupplierId((current) => current || items[0]?.id || '')
      })
      .catch((error) => {
        if (cancel) return
        if (isApiError(error) && (error.status === 401 || error.status === 403)) {
          setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
          return
        }
        setProblem(true)
        setMessage(isApiError(error) ? error.message : 'Suppliers could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [loadAdminSuppliers])

  function updateTier(index: number, key: keyof TierDraft, value: string) {
    setTiers((current) => current.map((tier, tierIndex) => (
      tierIndex === index ? { ...tier, [key]: value } : tier
    )))
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const nextErrors: Record<string, string> = {}
    if (!supplierId) nextErrors.supplierId = 'Choose a supplier'
    if (!name.trim()) nextErrors.name = 'Name is required'
    if (!category.trim()) nextErrors.category = 'Category is required'
    if (!unit.trim()) nextErrors.unit = 'Unit is required'
    if (retailPrice.trim() === '' || Number.isNaN(Number(retailPrice)) || Number(retailPrice) < 0) {
      nextErrors.retailPrice = 'Enter a retail price of 0 or more'
    }
    if (tiers.length === 0) nextErrors.priceTiers = 'Add at least one price tier'
    const seen = new Set<number>()
    let duplicate = false
    tiers.forEach((tier, index) => {
      const minQty = Number(tier.minQty)
      if (tier.minQty.trim() === '' || !Number.isInteger(minQty) || minQty < 1) {
        nextErrors[`priceTiers.${index}.minQty`] = 'Enter a minimum quantity of at least 1'
      } else if (seen.has(minQty)) {
        duplicate = true
      } else {
        seen.add(minQty)
      }
      const unitPrice = Number(tier.unitPrice)
      if (tier.unitPrice.trim() === '' || Number.isNaN(unitPrice) || unitPrice < 0) {
        nextErrors[`priceTiers.${index}.unitPrice`] = 'Enter a unit price of 0 or more'
      }
    })
    if (duplicate) nextErrors.priceTiers = 'Each minimum quantity can only be used once'
    setFieldErrors(nextErrors)
    setProblem(false)
    setMessage('')
    if (Object.keys(nextErrors).length > 0) return
    setPending(true)
    try {
      const imageUrl = photo ? await api.uploadPhoto(photo) : undefined
      const created = await createProduct({
        supplierId,
        name: name.trim(),
        description: description.trim(),
        category: category.trim(),
        tags: tags.split(',').map((tag) => tag.trim()).filter(Boolean),
        unit: unit.trim(),
        retailPrice: Number(retailPrice),
        imageUrl,
        priceTiers: tiers.map((tier) => ({ minQty: Number(tier.minQty), unitPrice: Number(tier.unitPrice) })),
      })
      setName('')
      setDescription('')
      setCategory('')
      setTags('')
      setUnit('')
      setRetailPrice('')
      setPhoto(null)
      setPreview('')
      setTiers([emptyTier()])
      setProblem(false)
      setMessage(`${created.name} was added.`)
    } catch (error) {
      if (isApiError(error) && (error.status === 401 || error.status === 403)) {
        setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
        return
      }
      setFieldErrors(fieldErrorsOf(error))
      setProblem(true)
      setMessage(isApiError(error) ? error.message : 'Could not add the product')
    } finally {
      setPending(false)
    }
  }

  if (blocked) {
    return <p role="alert" className="text-danger">{blocked}</p>
  }

  return (
    <Card className="max-w-2xl p-6">
      <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
        {message && (
          <p role={problem ? 'alert' : 'status'} className={problem ? 'text-sm text-danger' : 'text-sm text-ink-800'}>
            {message}
          </p>
        )}
        <SelectField id="product-supplier" label="Supplier" value={supplierId} error={fieldErrors.supplierId} onChange={(event) => setSupplierId(event.target.value)}>
          {suppliers.length === 0 && <option value="">No suppliers yet</option>}
          {suppliers.map((supplier) => (
            <option key={supplier.id} value={supplier.id}>
              {supplier.name}{supplier.isVerified ? '' : ' (not verified)'}
            </option>
          ))}
        </SelectField>
        <label className="block" htmlFor="product-photo">
          <span className="text-sm font-semibold text-ink-900">Photo</span>
          <input
            id="product-photo"
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="mt-1.5 block w-full text-sm text-ink-800"
            onChange={(event) => {
              const file = event.target.files?.[0] ?? null
              setPhoto(file)
              setPreview(file ? URL.createObjectURL(file) : '')
            }}
          />
          {fieldErrors.imageUrl && <span className="mt-1 block text-sm text-danger">{fieldErrors.imageUrl}</span>}
          {preview && <img src={mediaUrl(preview)} alt="" className="mt-3 aspect-[4/3] w-full rounded-md object-cover" />}
        </label>
        <TextField id="product-name" label="Name" value={name} error={fieldErrors.name} onChange={(event) => setName(event.target.value)} />
        <TextField id="product-description" label="Description" value={description} error={fieldErrors.description} onChange={(event) => setDescription(event.target.value)} />
        <TextField id="product-category" label="Category" value={category} error={fieldErrors.category} onChange={(event) => setCategory(event.target.value)} />
        <TextField id="product-tags" label="Tags" value={tags} placeholder="mailer, boxes" error={fieldErrors.tags} onChange={(event) => setTags(event.target.value)} />
        <TextField id="product-unit" label="Unit" value={unit} placeholder="carton" error={fieldErrors.unit} onChange={(event) => setUnit(event.target.value)} />
        <TextField id="product-retail" label="Retail price" type="number" min={0} step="0.01" value={retailPrice} error={fieldErrors.retailPrice} onChange={(event) => setRetailPrice(event.target.value)} />
        <fieldset className="grid gap-4">
          <legend className="text-sm font-semibold text-ink-900">Price tiers</legend>
          {fieldErrors.priceTiers && <p className="text-sm text-danger">{fieldErrors.priceTiers}</p>}
          {tiers.map((tier, index) => (
            <div key={index} className="grid gap-3 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
              <TextField
                id={`tier-min-${index}`}
                label={index === 0 ? 'Minimum quantity' : `Minimum quantity ${index + 1}`}
                type="number"
                min={1}
                step={1}
                value={tier.minQty}
                error={fieldErrors[`priceTiers.${index}.minQty`]}
                onChange={(event) => updateTier(index, 'minQty', event.target.value)}
              />
              <TextField
                id={`tier-price-${index}`}
                label="Unit price"
                type="number"
                min={0}
                step="0.01"
                value={tier.unitPrice}
                error={fieldErrors[`priceTiers.${index}.unitPrice`]}
                onChange={(event) => updateTier(index, 'unitPrice', event.target.value)}
              />
              {tiers.length > 1 && (
                <Button variant="ghost" className="mb-0.5" onClick={() => setTiers((current) => current.filter((_, tierIndex) => tierIndex !== index))}>
                  Remove
                </Button>
              )}
            </div>
          ))}
          <Button variant="secondary" onClick={() => setTiers((current) => [...current, emptyTier()])}>
            Add a tier
          </Button>
        </fieldset>
        <Button type="submit" disabled={pending}>{pending ? 'Adding…' : 'Add product'}</Button>
      </form>
    </Card>
  )
}
